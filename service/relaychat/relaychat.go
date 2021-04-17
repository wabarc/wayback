// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package relaychat // import "github.com/wabarc/wayback/service/relaychat"

import (
	"context"
	"crypto/tls"
	"os"
	"os/signal"
	"sync"
	"syscall"

	irc "github.com/thoj/go-ircevent"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/publish"
)

type IRC struct {
	sync.RWMutex

	conn *irc.Connection
}

// New IRC struct.
func New() *IRC {
	if config.Opts.IRCNick() == "" {
		logger.Fatal("Missing required environment variable")
	}

	// TODO: support SASL authenticate
	conn := irc.IRC(config.Opts.IRCNick(), config.Opts.IRCNick())
	conn.Password = config.Opts.IRCPassword()
	conn.VerboseCallbackHandler = config.Opts.HasDebugMode()
	conn.Debug = config.Opts.HasDebugMode()
	conn.UseTLS = true
	conn.TLSConfig = &tls.Config{InsecureSkipVerify: false}

	return &IRC{
		conn: conn,
	}
}

// Serve loop request direct messages from the IRC server.
// Serve returns an error.
func (i *IRC) Serve(ctx context.Context) error {
	if i.conn == nil {
		return errors.New("Must initialize IRC connection.")
	}
	logger.Debug("[irc] Serving IRC instance: %s", config.Opts.IRCServer())

	i.conn.AddCallback("PRIVMSG", func(ev *irc.Event) {
		go func(ev *irc.Event) {
			if err := i.process(context.Background(), ev); err != nil {
				logger.Error("[irc] Process failure, message: %s, error: %v", ev.Message(), err)
			}
		}(ev)
	})
	err := i.conn.Connect(config.Opts.IRCServer())
	if err != nil {
		logger.Error("[irc] Get conversations failure, error: %v", err)
		return err
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		i.conn.Quit()
	}()

	i.conn.Loop()
	return nil
}

func (i *IRC) process(ctx context.Context, ev *irc.Event) error {
	if ev.Nick == "" || ev.Message() == "" {
		logger.Debug("[irc] without nick or empty message")
		return errors.New("IRC: without nick or enpty message")
	}

	text := ev.MessageWithoutFormat()
	logger.Debug("[irc] from: %s message: %s", ev.Nick, text)

	urls := helper.MatchURL(text)
	if len(urls) == 0 {
		logger.Info("[irc] archives failure, URL no found.")
		return errors.New("IRC: URL no found")
	}

	col, err := i.archive(urls)
	if err != nil {
		logger.Error("[irc] archives failure, %v", err)
		return err
	}

	pub := publish.NewIRC(i.conn)
	replyText := pub.Render(col)

	// Reply result to sender
	i.conn.Privmsg(ev.Nick, replyText)

	// Reply and publish toot as public
	ctx = context.WithValue(ctx, "irc", i.conn)
	publish.To(ctx, col, "irc")

	return nil
}

func (i *IRC) archive(urls []string) (col []*wayback.Collect, err error) {
	logger.Debug("[irc] archives start...")

	wg := sync.WaitGroup{}
	var wbrc wayback.Broker = &wayback.Handle{URLs: urls}
	for slot, arc := range config.Opts.Slots() {
		if !arc {
			continue
		}
		wg.Add(1)
		go func(slot string) {
			defer wg.Done()
			c := &wayback.Collect{}
			logger.Debug("[irc] archiving slot: %s", slot)
			switch slot {
			case config.SLOT_IA:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.IA()
			case config.SLOT_IS:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.IS()
			case config.SLOT_IP:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.IP()
			case config.SLOT_PH:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.PH()
			}
			col = append(col, c)
		}(slot)
	}
	wg.Wait()

	if len(col) == 0 {
		logger.Error("archives failure")
		return col, errors.New("archives failure")
	}
	if len(col[0].Dst) == 0 {
		logger.Error("without results")
		return col, errors.New("without results")
	}

	return col, nil
}
