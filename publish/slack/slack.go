// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package slack // import "github.com/wabarc/wayback/publish/slack"

import (
	"context"
	"net/http"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/template/render"

	slack "github.com/slack-go/slack"
)

// Interface guard
var _ publish.Publisher = (*Slack)(nil)

type Slack struct {
	ctx context.Context

	bot  *slack.Client
	opts *config.Options
}

// New returns Slack bot client
func New(ctx context.Context, client *http.Client, opts *config.Options) *Slack {
	if !opts.PublishToSlackChannel() {
		logger.Debug("Missing required environment variable, abort.")
		return nil
	}

	options := []slack.Option{slack.OptionDebug(opts.HasDebugMode())}
	if client != nil {
		options = append(options, slack.OptionHTTPClient(client))
	}
	bot := slack.New(
		opts.SlackBotToken(),
		options...,
	)
	if bot == nil {
		logger.Fatal("create slack bot instance failed")
		return nil
	}

	return &Slack{ctx: ctx, bot: bot, opts: opts}
}

// Publish publish text to the Slack channel of given cols and args.
// A context should contains a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (s *Slack) Publish(ctx context.Context, rdx reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusFailure)
		return errors.New("publish to slack: collects empty")
	}

	art, err := publish.Artifact(ctx, rdx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	var head = render.Title(cols, rdx)
	var body = render.ForPublish(&render.Slack{Cols: cols, Data: rdx}).String()
	if s.toChannel(art, head, body) {
		metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusSuccess)
		return nil
	}
	metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusFailure)
	return errors.New("publish to slack failed")
}

// toChannel for publish to message to Slack channel,
// returns boolean as result.
func (s *Slack) toChannel(art reduxer.Artifact, head, body string) (ok bool) {
	if body == "" {
		logger.Warn("post to message to channel failed, body empty")
		return ok
	}

	msgOpts := []slack.MsgOption{
		slack.MsgOptionText(body, false),
		slack.MsgOptionDisableMarkdown(),
	}
	_, tstamp, err := s.bot.PostMessage(s.opts.SlackChannel(), msgOpts...)
	if err != nil {
		logger.Error("post message failed: %v", err)
		return false
	}
	if err := service.UploadToSlack(s.bot, s.opts, art, s.opts.SlackChannel(), tstamp, head); err != nil {
		logger.Error("upload files to slack failed: %v", err)
	}

	return true
}

// Shutdown shuts down the Slack publish service, it always return a nil error.
func (s *Slack) Shutdown() error {
	return nil
}
