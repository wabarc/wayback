// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package relaychat // import "github.com/wabarc/wayback/publish/relaychat"

import (
	"context"
	"crypto/tls"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"

	irc "github.com/thoj/go-ircevent"
)

// Interface guard
var _ publish.Publisher = (*IRC)(nil)

type IRC struct {
	conn *irc.Connection
	opts *config.Options
}

// New returns a IRC struct
func New(conn *irc.Connection, opts *config.Options) *IRC {
	if !opts.PublishToIRCChannel() {
		logger.Debug("Missing required environment variable, abort.")
		return nil
	}

	if conn == nil {
		conn = irc.IRC(opts.IRCNick(), opts.IRCNick())
		conn.Password = opts.IRCPassword()
		conn.VerboseCallbackHandler = opts.HasDebugMode()
		conn.Debug = opts.HasDebugMode()
		conn.UseTLS = true
		conn.TLSConfig = &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12}
	}

	return &IRC{conn: conn, opts: opts}
}

// Publish publish text to IRC channel of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (i *IRC) Publish(ctx context.Context, _ reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusFailure)
		return errors.New("publish to irc: collects empty")
	}

	var txt = render.ForPublish(&render.Relaychat{Cols: cols}).String()
	if i.toChannel(ctx, txt) {
		metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusSuccess)
		return nil
	}
	metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusFailure)
	return errors.New("publish to irc failed")
}

func (i *IRC) toChannel(_ context.Context, text string) bool {
	if !i.opts.PublishToIRCChannel() || i.conn == nil {
		logger.Warn("Do not publish to IRC channel.")
		return false
	}
	if text == "" {
		logger.Warn("IRC validation failed: Text can't be blank")
		return false
	}

	go func() {
		// i.conn.Join(o.opts.IRCChannel())
		i.conn.Privmsg(i.opts.IRCChannel(), text)
	}()

	return true
}

// Shutdown shuts down the IRC publish service.
func (i *IRC) Shutdown() error {
	if i.conn != nil {
		i.conn.Quit()
	}

	return nil
}
