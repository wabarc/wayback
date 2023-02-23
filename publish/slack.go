// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/template/render"

	slack "github.com/slack-go/slack"
)

var _ Publisher = (*slackBot)(nil)

type slackBot struct {
	bot  *slack.Client
	opts *config.Options
}

// NewSlack returns Slack bot client
func NewSlack(bot *slack.Client, opts *config.Options) *slackBot {
	if !opts.PublishToSlackChannel() {
		logger.Error("Missing required environment variable, abort.")
		return new(slackBot)
	}

	if bot == nil {
		bot = slack.New(
			opts.SlackBotToken(),
			slack.OptionDebug(opts.HasDebugMode()),
		)
		if bot == nil {
			logger.Error("create slack bot instance failed")
		}
	}

	return &slackBot{bot: bot, opts: opts}
}

// Publish publish text to the Slack channel of given cols and args.
// A context should contains a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (s *slackBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusRequest)

	if len(cols) == 0 {
		return errors.New("publish to slack: collects empty")
	}

	rdx, art, err := extract(ctx, cols)
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
func (s *slackBot) toChannel(art reduxer.Artifact, head, body string) (ok bool) {
	if body == "" {
		logger.Warn("post to message to channel failed, body empty")
		return ok
	}
	if s.bot == nil {
		s.bot = slack.New(s.opts.SlackBotToken())
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
