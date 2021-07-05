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
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/template/render"
)

type IRC struct {
	sync.RWMutex

	ctx   context.Context
	pool  pooling.Pool
	conn  *irc.Connection
	store *storage.Storage
}

// New IRC struct.
func New(ctx context.Context, store *storage.Storage, pool pooling.Pool) *IRC {
	if config.Opts.IRCNick() == "" {
		logger.Fatal("[irc] missing required environment variable")
	}
	if store == nil {
		logger.Fatal("[irc] must initialize storage")
	}
	if pool == nil {
		logger.Fatal("[irc] must initialize pooling")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	// TODO: support SASL authenticate
	conn := irc.IRC(config.Opts.IRCNick(), config.Opts.IRCNick())
	conn.Password = config.Opts.IRCPassword()
	conn.VerboseCallbackHandler = config.Opts.HasDebugMode()
	conn.Debug = config.Opts.HasDebugMode()
	conn.UseTLS = true
	conn.TLSConfig = &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12}

	return &IRC{
		ctx:   ctx,
		pool:  pool,
		conn:  conn,
		store: store,
	}
}

// Serve loop request direct messages from the IRC server.
// Serve returns an error.
func (i *IRC) Serve() error {
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
			go i.pool.Roll(func() {
				if err := i.process(ev); err != nil {
					logger.Error("[irc] process failure, message: %s, error: %v", ev.Message(), err)
					metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusFailure)
				} else {
					metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusSuccess)
				}
			})
		}(ev)
	})
	err := i.conn.Connect(config.Opts.IRCServer())
	if err != nil {
		logger.Error("[irc] Get conversations failure, error: %v", err)
		return err
	}

	go func() {
		<-i.ctx.Done()
		i.conn.Quit()
	}()

	i.conn.Loop()
	return nil
}

func (i *IRC) process(ev *irc.Event) error {
	if ev.Nick == "" || ev.Message() == "" {
		logger.Debug("[irc] without nick or empty message")
		return errors.New("IRC: without nick or enpty message")
	}

	text := ev.MessageWithoutFormat()
	logger.Debug("[irc] from: %s message: %s", ev.Nick, text)

	urls := helper.MatchURLFallback(text)
	if len(urls) == 0 {
		logger.Info("[irc] archives failure, URL no found.")
		return errors.New("IRC: URL no found")
	}

	var bundles reduxer.Bundles
	cols, err := wayback.Wayback(context.TODO(), &bundles, urls...)
	if err != nil {
		logger.Error("[irc] archives failure, %v", err)
		return err
	}
	logger.Debug("[irc] bundles: %#v", bundles)

	replyText := render.ForReply(&render.Relaychat{Cols: cols}).String()

	// Reply result to sender
	i.conn.Privmsg(ev.Nick, replyText)

	// Reply and publish toot as public
	ctx := context.WithValue(i.ctx, publish.FlagIRC, i.conn)
	ctx = context.WithValue(ctx, publish.PubBundle, bundles)
	publish.To(ctx, cols, publish.FlagIRC)

	return nil
}
