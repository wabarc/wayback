// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package relaychat // import "github.com/wabarc/wayback/service/relaychat"

import (
	"context"
	"crypto/tls"
	"os"
	"strings"
	"testing"
	"time"

	irc "github.com/thoj/go-ircevent"
	"github.com/wabarc/wayback/config"
)

const (
	server1  = "irc.freenode.net:7000"
	server2  = "irc.darkscience.net:6697"
	sender   = "wabarc-sender"
	receiver = "wabarc-receiver"
	channel  = "#wabarc-testing"
	debug    = false
)

func conn(nick string) *irc.Connection {
	i := irc.IRC(nick, nick)
	i.UseTLS = true
	i.VerboseCallbackHandler = debug
	i.Debug = debug
	i.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return i
}

// Bash: echo -e 'USER wabarc-sender guest * *\nNICK wabarc-sender\nPRIVMSG wabarc-receiver :Hello, World!\nQUIT\n' \ | nc irc.freenode.net 6667
func TestProcess(t *testing.T) {
	os.Setenv("WAYBACK_IRC_NICK", "wabarc-process")
	os.Setenv("WAYBACK_IRC_SERVER", server1)
	os.Setenv("WAYBACK_IRC_CHANNEL", channel)
	os.Setenv("WAYBACK_ENABLE_IA", "true")

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse enviroment variables or flags failed, error: %v", err)
	}

	sendConn := conn(sender)
	recvConn := conn(receiver)
	msgok := make(chan bool, 1)

	// Send privmsg
	go func() {
		tick := time.NewTicker(5 * time.Second)
		i := 1
		for {
			select {
			case <-tick.C:
				if i == 1 {
					sendConn.Privmsg(receiver, "privmsg from sender https://example.com")
					break
				}
			case <-msgok:
				tick.Stop()
				break
			}
			i -= 1
		}
	}()

	// Join channel and receive message
	// sendConn.AddCallback("001", func(ev *irc.Event) {sendConn.Join(channel) })
	// Receive response from receiver
	sendConn.AddCallback("PRIVMSG", func(ev *irc.Event) {
		if ev.Nick == receiver {
			t.Log("message: " + ev.Message())
			if !strings.Contains(ev.Message(), config.SlotName("ia")) {
				t.Fail()
			}
			sendConn.Quit()
		}
	})

	// Receive privmsg from sender
	recvConn.AddCallback("PRIVMSG", func(ev *irc.Event) {
		t.Log("from: " + ev.Nick)
		t.Log("message: " + ev.Message())
		if ev.Nick == sender {
			i := New(config.Opts)
			// Replace IRC connection to receive connection
			i.conn = recvConn
			if err = i.process(context.Background(), ev); err != nil {
				t.Error(err)
			}
			msgok <- true
			recvConn.Quit()
		}
	})

	err = sendConn.Connect(server1)
	if err != nil {
		t.Log(err.Error())
		t.Errorf("Can't connect to freenode.")
	}
	err = recvConn.Connect(server1)
	if err != nil {
		t.Log(err.Error())
		t.Errorf("Can't connect to freenode.")
	}

	go recvConn.Loop()
	sendConn.Loop()
}
