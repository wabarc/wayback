// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"crypto/tls"

	irc "github.com/thoj/go-ircevent"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/template/render"
)

type ircBot struct {
	conn *irc.Connection
}

// NewIRC returns a ircBot struct
func NewIRC(conn *irc.Connection) *ircBot {
	if !config.Opts.PublishToIRCChannel() {
		logger.Error("Missing required environment variable, abort.")
		return new(ircBot)
	}

	if conn == nil {
		conn = irc.IRC(config.Opts.IRCNick(), config.Opts.IRCNick())
		conn.Password = config.Opts.IRCPassword()
		conn.VerboseCallbackHandler = config.Opts.HasDebugMode()
		conn.Debug = config.Opts.HasDebugMode()
		conn.UseTLS = true
		conn.TLSConfig = &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12}
	}

	return &ircBot{conn: conn}
}

// Publish publish text to IRC channel of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` constant.
func (i *ircBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Warn("collects empty")
		return
	}

	var txt = render.ForPublish(&render.Relaychat{Cols: cols}).String()
	if i.toChannel(ctx, txt) {
		metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusFailure)
	return
}

func (i *ircBot) toChannel(_ context.Context, text string) bool {
	if !config.Opts.PublishToIRCChannel() || i.conn == nil {
		logger.Warn("Do not publish to IRC channel.")
		return false
	}
	if text == "" {
		logger.Warn("IRC validation failed: Text can't be blank")
		return false
	}

	go func() {
		// i.conn.Join(config.Opts.IRCChannel())
		i.conn.Privmsg(config.Opts.IRCChannel(), text)
	}()

	return true
}
