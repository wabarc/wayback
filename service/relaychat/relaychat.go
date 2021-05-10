// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package relaychat // import "github.com/wabarc/wayback/service/relaychat"

import (
	"context"
	"crypto/tls"
	"sync"

	irc "github.com/thoj/go-ircevent"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/storage"
)

type IRC struct {
	sync.RWMutex

	conn  *irc.Connection
	store *storage.Storage
}

// New IRC struct.
func New(store *storage.Storage) *IRC {
	if config.Opts.IRCNick() == "" {
		logger.Fatal("[irc] missing required environment variable")
	}
	if store == nil {
		logger.Fatal("[irc] must initialize storage")
	}

	// TODO: support SASL authenticate
	conn := irc.IRC(config.Opts.IRCNick(), config.Opts.IRCNick())
	conn.Password = config.Opts.IRCPassword()
	conn.VerboseCallbackHandler = config.Opts.HasDebugMode()
	conn.Debug = config.Opts.HasDebugMode()
	conn.UseTLS = true
	conn.TLSConfig = &tls.Config{InsecureSkipVerify: false}

	return &IRC{
		conn:  conn,
		store: store,
	}
}

// Serve loop request direct messages from the IRC server.
// Serve returns an error.
func (i *IRC) Serve(ctx context.Context) error {
	if i.conn == nil {
		return errors.New("Must initialize IRC connection.")
	}
	logger.Debug("[irc] Serving IRC instance: %s", config.Opts.IRCServer())

	if config.Opts.IRCChannel() != "" {
		i.conn.AddCallback("001", func(ev *irc.Event) { i.conn.Join(config.Opts.IRCChannel()) })
	}
	i.conn.AddCallback("PRIVMSG", func(ev *irc.Event) {
		go func(ev *irc.Event) {
			metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusRequest)
			if err := i.process(context.Background(), ev); err != nil {
				logger.Error("[irc] process failure, message: %s, error: %v", ev.Message(), err)
				metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusFailure)
			} else {
				metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusSuccess)
			}
		}(ev)
	})
	err := i.conn.Connect(config.Opts.IRCServer())
	if err != nil {
		logger.Error("[irc] Get conversations failure, error: %v", err)
		return err
	}

	go func() {
		select {
		case <-ctx.Done():
			i.conn.Quit()
		}
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

	col, err := wayback.Wayback(urls)
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
