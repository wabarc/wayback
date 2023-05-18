// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package relaychat // import "github.com/wabarc/wayback/publish/relaychat"

import (
	"context"
	"strings"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"
	"gopkg.in/irc.v4"
)

// Interface guard
var _ publish.Publisher = (*IRC)(nil)

type IRC struct {
	conn *irc.Client
	opts *config.Options
}

// New returns a IRC struct
func New(c *irc.Client, opts *config.Options) *IRC {
	if !opts.PublishToIRCChannel() {
		logger.Debug("Missing required environment variable, abort.")
		return nil
	}

	return &IRC{opts: opts}
}

// Publish publish text to IRC channel of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (i *IRC) Publish(ctx context.Context, _ reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	// Most IRC server supports establish one connection,
	// this value accessed from service module.
	if i.conn == nil {
		v := ctx.Value(publish.FlagIRC)
		conn, ok := v.(*irc.Client)
		if ok {
			i.conn = conn
		}
	}

	metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusFailure)
		return errors.New("publish to irc: collects empty")
	}

	txt := strings.Split(render.ForPublish(&render.Relaychat{Cols: cols}).String(), "\n")
	if i.toChannel(ctx, txt...) {
		metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusSuccess)
		return nil
	}
	metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusFailure)
	return errors.New("publish to irc failed")
}

func (i *IRC) toChannel(_ context.Context, text ...string) bool {
	if !i.opts.PublishToIRCChannel() || i.conn == nil {
		logger.Warn("Do not publish to IRC channel.")
		return false
	}
	if len(text) == 0 {
		logger.Warn("IRC validation failed: Text can't be blank")
		return false
	}

	err := i.reply(i.opts.IRCChannel(), text...)
	if err != nil {
		logger.Error("publish to IRC channel failed: %v", err)
		return false
	}

	return true
}

// Shutdown shuts down the IRC publish service.
func (i *IRC) Shutdown() error {
	return nil
}

func (i *IRC) reply(name string, messages ...string) (err error) {
	if i.conn == nil {
		return errors.New("irc connection is missing")
	}

	for _, text := range messages {
		text = strings.ReplaceAll(text, "\n", " ")
		msg := &irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				name,
				text,
			},
		}
		if e := i.conn.WriteMessage(msg); e != nil {
			err = errors.Wrap(err, e.Error())
		}
	}
	return err
}
