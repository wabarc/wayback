// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"os"

	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/reduxer"
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
// A context should contains a `reduxer.Bundle` via `publish.PubBundle` constant.
func (s *slackBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Warn("collects empty")
		return
	}

	var bnd = bundle(ctx, cols)
	var txt = render.ForPublish(&render.Slack{Cols: cols, Data: bnd}).String()
	if s.toChannel(bnd, txt) {
		metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusFailure)
	return
}

// toChannel for publish to message to Slack channel,
// returns boolean as result.
func (s *slackBot) toChannel(bundle *reduxer.Bundle, text string) (ok bool) {
	if text == "" {
		logger.Warn("post to message to channel failed, text empty")
		return ok
	}
	if s.bot == nil {
		s.bot = slack.New(config.Opts.SlackBotToken())
	}

	msgOpts := []slack.MsgOption{
		slack.MsgOptionText(text, false),
		slack.MsgOptionDisableMarkdown(),
	}
	_, tstamp, err := s.bot.PostMessage(config.Opts.SlackChannel(), msgOpts...)
	if err != nil {
		logger.Error("post message failed: %v", err)
		return false
	}
	if err := UploadToSlack(s.bot, bundle, config.Opts.SlackChannel(), tstamp); err != nil {
		logger.Error("upload files to slack failed: %v", err)
	}

	return true
}

// UploadToSlack upload files to channel and attach as a reply by the given bundle
func UploadToSlack(client *slack.Client, bundle *reduxer.Bundle, channel, timestamp string) (err error) {
	if client == nil {
		return errors.New("client invalid")
	}

	var fsize int64
	for _, asset := range bundle.Asset() {
		if asset.Local == "" {
			continue
		}
		if !helper.Exists(asset.Local) {
			err = errors.Wrap(err, "invalid file "+asset.Local)
			continue
		}
		fsize += helper.FileSize(asset.Local)
		if fsize > config.Opts.MaxAttachSize("slack") {
			err = errors.Wrap(err, "total file size large than 5GB, skipped")
			continue
		}
		reader, e := os.Open(asset.Local)
		if e != nil {
			err = errors.Wrap(err, e.Error())
			continue
		}
		params := slack.FileUploadParameters{
			Filename:        asset.Local,
			Reader:          reader,
			Title:           bundle.Title,
			Channels:        []string{channel},
			ThreadTimestamp: timestamp,
		}
		file, e := client.UploadFile(params)
		if e != nil {
			err = errors.Wrap(err, e.Error())
			continue
		}
		file, _, _, e = client.ShareFilePublicURL(file.ID)
		if e != nil {
			err = errors.Wrap(err, e.Error())
			continue
		}
		logger.Info("slack external file permalink: %s", file.PermalinkPublic)
	}

	return nil
}
