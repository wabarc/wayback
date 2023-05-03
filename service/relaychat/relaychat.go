// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package relaychat // import "github.com/wabarc/wayback/service/relaychat"

import (
	"context"
	"crypto/tls"
	"sync"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/template/render"

	irc "github.com/thoj/go-ircevent"
)

// Interface guard
var _ service.Servicer = (*IRC)(nil)

// ErrServiceClosed is returned by the Service's Serve method after a call to Shutdown.
var ErrServiceClosed = errors.New("irc: Service closed")

// IRC represents an IRC service in the application.
type IRC struct {
	sync.RWMutex

	ctx   context.Context
	opts  *config.Options
	pool  *pooling.Pool
	conn  *irc.Connection
	store *storage.Storage
	pub   *publish.Publish
}

// New IRC struct.
func New(ctx context.Context, opts service.Options) (*IRC, error) {
	if !opts.Config.IRCEnabled() {
		return nil, errors.New("missing required environment variable, skipped")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	// TODO: support SASL authenticate
	conn := irc.IRC(opts.Config.IRCNick(), opts.Config.IRCNick())
	conn.Password = opts.Config.IRCPassword()
	conn.VerboseCallbackHandler = opts.Config.HasDebugMode()
	conn.Debug = opts.Config.HasDebugMode()
	conn.UseTLS = true
	conn.TLSConfig = &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12}

	return &IRC{
		ctx:   ctx,
		conn:  conn,
		store: opts.Storage,
		opts:  opts.Config,
		pool:  opts.Pool,
		pub:   opts.Publish,
	}, nil
}

// Serve loop request direct messages from the IRC server.
// Serve returns an error.
func (i *IRC) Serve() error {
	if i.conn == nil {
		return errors.New("Must initialize IRC connection.")
	}
	logger.Info("Serving IRC instance: %s", i.opts.IRCServer())

	if i.opts.IRCChannel() != "" {
		i.conn.AddCallback("001", func(ev *irc.Event) { i.conn.Join(i.opts.IRCChannel()) })
	}
	i.conn.AddCallback("PRIVMSG", func(ev *irc.Event) {
		go func(ev *irc.Event) {
			metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusRequest)
			bucket := pooling.Bucket{
				Request: func(ctx context.Context) error {
					if err := i.process(ctx, ev); err != nil {
						logger.Error("process failure, message: %s, error: %v", ev.Message(), err)
						return err
					}
					metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusSuccess)
					return nil
				},
				Fallback: func(_ context.Context) error {
					i.conn.Privmsg(ev.Nick, service.MsgWaybackTimeout)
					metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusFailure)
					return nil
				},
			}
			i.pool.Put(bucket)
		}(ev)
	})
	err := i.conn.Connect(i.opts.IRCServer())
	if err != nil {
		logger.Error("Get conversations failure, error: %v", err)
		return err
	}

	go func() {
		i.conn.Loop()
	}()

	// Block until context cone
	<-i.ctx.Done()

	return ErrServiceClosed
}

// Shutdown shuts down the IRC service, it always retuan a nil error.
func (i *IRC) Shutdown() error {
	if i.conn != nil {
		i.conn.Quit()
	}

	return nil
}

func (i *IRC) process(ctx context.Context, ev *irc.Event) error {
	if ev.Nick == "" || ev.Message() == "" {
		logger.Warn("without nick or empty message")
		return errors.New("IRC: without nick or enpty message")
	}

	text := ev.MessageWithoutFormat()
	logger.Debug("from: %s message: %s", ev.Nick, text)

	urls := service.MatchURL(i.opts, text)
	if len(urls) == 0 {
		logger.Warn("archives failure, URL no found.")
		return errors.New("IRC: URL no found")
	}

	do := func(cols []wayback.Collect, rdx reduxer.Reduxer) error {
		logger.Debug("reduxer: %#v", rdx)

		replyText := render.ForReply(&render.Relaychat{Cols: cols}).String()

		// Reply result to sender
		i.conn.Privmsg(ev.Nick, replyText)

		// Reply and publish toot as public
		i.pub.Spread(ctx, rdx, cols, publish.FlagIRC)
		return nil
	}

	return service.Wayback(ctx, i.opts, urls, do)
}
