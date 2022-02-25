// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/template/render"

	slack "github.com/slack-go/slack"
)

type slackBot struct {
	bot *slack.Client
}

// NewSlack returns Slack bot client
func NewSlack(bot *slack.Client) *slackBot {
	if !config.Opts.PublishToSlackChannel() {
		logger.Error("Missing required environment variable, abort.")
		return new(slackBot)
	}

	if bot == nil {
		bot = slack.New(
			config.Opts.SlackBotToken(),
			slack.OptionDebug(config.Opts.HasDebugMode()),
		)
		if bot == nil {
			logger.Error("create slack bot instance failed")
		}
	}

	return &slackBot{bot: bot}
}

// Publish publish text to the Slack channel of given cols and args.
// A context should contains a `reduxer.Reduxer` via `publish.PubBundle` constant.
func (s *slackBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Warn("collects empty")
		return
	}

	rdx, art, err := extract(ctx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	var head = render.Title(cols, rdx)
	var body = render.ForPublish(&render.Slack{Cols: cols, Data: rdx}).String()
	if s.toChannel(art, head, body) {
		metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusFailure)
	return
}

// toChannel for publish to message to Slack channel,
// returns boolean as result.
func (s *slackBot) toChannel(art reduxer.Artifact, head, body string) (ok bool) {
	if body == "" {
		logger.Warn("post to message to channel failed, body empty")
		return ok
	}
	if s.bot == nil {
		s.bot = slack.New(config.Opts.SlackBotToken())
	}

	msgOpts := []slack.MsgOption{
		slack.MsgOptionText(body, false),
		slack.MsgOptionDisableMarkdown(),
	}
	_, tstamp, err := s.bot.PostMessage(config.Opts.SlackChannel(), msgOpts...)
	if err != nil {
		logger.Error("post message failed: %v", err)
		return false
	}
	if err := service.UploadToSlack(s.bot, art, config.Opts.SlackChannel(), tstamp, head); err != nil {
		logger.Error("upload files to slack failed: %v", err)
	}

	return true
}
