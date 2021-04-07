// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

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
)

var (
	server1  = "irc.freenode.net:7000"
	server2  = "irc.darkscience.net:6697"
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
	os.Setenv("WAYBACK_IRC_SERVER", server1)
	os.Setenv("WAYBACK_IRC_CHANNEL", channel)
	os.Setenv("WAYBACK_ENABLE_IA", "true")

	var err error
	var opts *config.Options
	parser := config.NewParser()
	if opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse enviroment variables or flags failed, error: %v", err)
	}

	sendConn := conn(sender)
	recvConn := conn(receiver)
	done := make(chan bool, 1)

	// Send privmsg if receiver connected
	recvConn.AddCallback("001", func(ev *irc.Event) {
		go func(ev *irc.Event) {
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
		}(ev)
	})

	// Receive privmsg from sender
	recvConn.AddCallback("PRIVMSG", func(ev *irc.Event) {
		if ev.Nick == sender {
			done <- true
			i := New(opts)
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

	err = recvConn.Connect(server1)
	if err != nil {
		t.Errorf("Can't connect to freenode, error: %v", err)
	}
	err = sendConn.Connect(server1)
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
	os.Setenv("WAYBACK_IRC_SERVER", server1)
	os.Setenv("WAYBACK_IRC_CHANNEL", channel)
	os.Setenv("WAYBACK_ENABLE_IA", "true")

	var err error
	var opts *config.Options
	parser := config.NewParser()
	if opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse enviroment variables or flags failed, error: %v", err)
	}

	sendConn := conn(sender)
	recvConn := conn(receiver)
	done := make(chan bool, 1)

	sendConn.AddCallback("001", func(ev *irc.Event) { sendConn.Join(channel) })
	recvConn.AddCallback("001", func(ev *irc.Event) { recvConn.Join(channel) })

	// Send privmsg if receiver connected
	recvConn.AddCallback("001", func(ev *irc.Event) {
		go func(ev *irc.Event) {
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
		}(ev)
	})

	// Receive privmsg from sender
	recvConn.AddCallback("PRIVMSG", func(ev *irc.Event) {
		if ev.Nick == sender {
			done <- true
			i := New(opts)
			// Replace IRC connection to receive connection
			i.conn = recvConn
			if err = i.process(context.Background(), ev); err != nil {
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

	err = recvConn.Connect(server1)
	if err != nil {
		t.Errorf("Can't connect to freenode, error: %v", err)
	}
	err = sendConn.Connect(server1)
	if err != nil {
		t.Errorf("Can't connect to freenode, error: %v", err)
	}

	go recvConn.Loop()
	sendConn.Loop()
}
