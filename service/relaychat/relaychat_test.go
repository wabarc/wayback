// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

//go:build !race
// +build !race

package relaychat // import "github.com/wabarc/wayback/service/relaychat"

import (
	"context"
	"crypto/tls"
	"os"
	"strings"
	"testing"
	"time"

	irc "github.com/thoj/go-ircevent"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"
)

var (
	server   = "irc.libera.chat:6697"
	sender   = "wsend" + helper.RandString(4, "lower")
	receiver = "wrecv" + helper.RandString(4, "lower")
	channel  = "#wabarc-testing"
	debug    = false
)

func conn(nick string) *irc.Connection {
	i := irc.IRC(nick, nick)
	i.UseTLS = true
	i.VerboseCallbackHandler = debug
	i.Debug = debug
	i.TLSConfig = &tls.Config{InsecureSkipVerify: false}

	return i
}

// Bash: echo -e 'USER wabarc-sender guest * *\nNICK wabarc-sender\nPRIVMSG wabarc-receiver :Hello, World!\nQUIT\n' \ | nc irc.freenode.net 6667
func TestProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test in short mode.")
	}

	os.Setenv("WAYBACK_IRC_NICK", "wabarc-process")
	os.Setenv("WAYBACK_IRC_SERVER", server)
	os.Setenv("WAYBACK_IRC_CHANNEL", channel)
	os.Setenv("WAYBACK_ENABLE_IA", "true")

	parser := config.NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	sendConn := conn(sender)
	recvConn := conn(receiver)
	done := make(chan bool, 1)

	// Send privmsg if receiver connected
	recvConn.AddCallback("001", func(ev *irc.Event) {
		go func() {
			tick := time.NewTicker(3 * time.Second)
			i := 10
			for {
				select {
				case <-tick.C:
					sendConn.Privmsg(receiver, "privmsg from sender https://example.com")
					if i == 0 {
						t.Errorf("Timeout while wating for test message from the other thread.")
						recvConn.Quit()
						sendConn.Quit()
						return
					}
				case <-done:
					tick.Stop()
				}
				i -= 1
			}
		}()
	})

	cfg := []pooling.Option{
		pooling.Capacity(opts.PoolingSize()),
		pooling.Timeout(opts.WaybackTimeout()),
		pooling.MaxRetries(opts.WaybackMaxRetries()),
	}
	ctx := context.Background()
	pool := pooling.New(ctx, cfg...)
	go pool.Roll()
	defer pool.Close()

	pub := publish.New(ctx, opts)
	defer pub.Stop()

	o := service.ParseOptions(service.Config(opts), service.Storage(&storage.Storage{}), service.Pool(pool), service.Publish(pub))
	// Receive privmsg from sender
	recvConn.AddCallback("PRIVMSG", func(ev *irc.Event) {
		if ev.Nick == sender {
			done <- true
			i, _ := New(context.Background(), o)
			// Replace IRC connection to receive connection
			i.conn = recvConn
			if err = i.process(context.Background(), ev); err != nil {
				t.Error(err)
			}
			recvConn.Quit()
		}
	})

	// Receive response from receiver
	sendConn.AddCallback("PRIVMSG", func(ev *irc.Event) {
		if ev.Nick == receiver {
			if !strings.Contains(ev.Message(), config.SlotName("ia")) {
				t.Errorf("Unexpected message: %s", ev.Message())
			}
			sendConn.Quit()
		}
	})

	err = recvConn.Connect(server)
	if err != nil {
		t.Errorf("Can't connect to freenode, error: %v", err)
	}
	err = sendConn.Connect(server)
	if err != nil {
		t.Errorf("Can't connect to freenode, error: %v", err)
	}

	go recvConn.Loop()
	sendConn.Loop()
}

func TestToIRCChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test in short mode.")
	}

	conn := func(nick string) *irc.Connection {
		i := irc.IRC(nick, nick)
		i.UseTLS = true
		i.VerboseCallbackHandler = debug
		i.Debug = debug
		i.TLSConfig = &tls.Config{InsecureSkipVerify: false}
		return i
	}

	os.Setenv("WAYBACK_IRC_NICK", "wabarc-process")
	os.Setenv("WAYBACK_IRC_SERVER", server)
	os.Setenv("WAYBACK_IRC_CHANNEL", channel)
	os.Setenv("WAYBACK_ENABLE_IA", "true")

	parser := config.NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	sendConn := conn(sender)
	recvConn := conn(receiver)
	done := make(chan bool, 1)

	sendConn.AddCallback("001", func(ev *irc.Event) { sendConn.Join(channel) })
	recvConn.AddCallback("001", func(ev *irc.Event) { recvConn.Join(channel) })

	// Send privmsg if receiver connected
	recvConn.AddCallback("001", func(ev *irc.Event) {
		go func() {
			tick := time.NewTicker(3 * time.Second)
			i := 10
			for {
				select {
				case <-tick.C:
					sendConn.Privmsg(receiver, "privmsg from sender https://example.com")
					if i == 0 {
						t.Errorf("Timeout while wating for test message from the other thread.")
						recvConn.Quit()
						sendConn.Quit()
						return
					}
				case <-done:
					tick.Stop()
				}
				i -= 1
			}
		}()
	})

	cfg := []pooling.Option{
		pooling.Capacity(opts.PoolingSize()),
		pooling.Timeout(opts.WaybackTimeout()),
		pooling.MaxRetries(opts.WaybackMaxRetries()),
	}
	ctx := context.Background()
	pool := pooling.New(ctx, cfg...)
	go pool.Roll()
	defer pool.Close()

	pub := publish.New(ctx, opts)
	defer pub.Stop()

	o := service.ParseOptions(service.Config(opts), service.Storage(&storage.Storage{}), service.Pool(pool), service.Publish(pub))
	// Receive privmsg from sender
	recvConn.AddCallback("PRIVMSG", func(ev *irc.Event) {
		if ev.Nick == sender {
			done <- true
			ctx := context.Background()
			i, _ := New(ctx, o)
			// Replace IRC connection to receive connection
			i.conn = recvConn
			if err = i.process(ctx, ev); err != nil {
				t.Error(err)
			}
			recvConn.Quit()
		}
	})

	// Receive response from channel
	sendConn.AddCallback("PRIVMSG", func(ev *irc.Event) {
		if len(ev.Arguments) == 0 {
			t.Fatal("Unexpected got IRC event")
		}
		if ev.Arguments[0] == channel {
			if !strings.Contains(ev.Message(), config.SlotName("ia")) {
				t.Errorf("Unexpected message: %s", ev.Message())
			}
			sendConn.Quit()
		}
	})

	err = recvConn.Connect(server)
	if err != nil {
		t.Errorf("Can't connect to freenode, error: %v", err)
	}
	err = sendConn.Connect(server)
	if err != nil {
		t.Errorf("Can't connect to freenode, error: %v", err)
	}

	go recvConn.Loop()
	sendConn.Loop()
}
