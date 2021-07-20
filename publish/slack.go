// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"os"
	"strings"

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

func (s *slackBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Warn("collects empty")
		return
	}

	var bnd = bundle(ctx, cols)
	var txt = render.ForPublish(&render.Slack{Cols: cols}).String()
	if s.toChannel(ctx, bnd, txt) {
		metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishSlack, metrics.StatusFailure)
	return
}

// toChannel for publish to message to Slack channel,
// returns boolean as result.
func (s *slackBot) toChannel(ctx context.Context, bundle *reduxer.Bundle, text string) (ok bool) {
	if text == "" {
		logger.Warn("post to message to channel failed, text empty")
		return ok
	}
	if s.bot == nil {
		s.bot = slack.New(config.Opts.SlackBotToken())
	}

	// TODO: move to render
	var b strings.Builder
	if head := title(ctx, bundle); head != "" {
		b.WriteString(`‹ `)
		b.WriteString(head)
		b.WriteString(" ›\n\n")
	}
	if dgst := digest(ctx, bundle); dgst != "" {
		b.WriteString(dgst)
		b.WriteString("\n\n")
	}
	b.WriteString(text)

	msgOpts := []slack.MsgOption{
		slack.MsgOptionText(b.String(), false),
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
func UploadToSlack(client *slack.Client, bundle *reduxer.Bundle, channel, timestamp string) error {
	if client == nil {
		return errors.New("client invalid")
	}

	var fsize int64
	// TODO: clean code and wrap errors
	for _, path := range bundle.Paths() {
		if path == "" {
			continue
		}
		if !helper.Exists(path) {
			logger.Warn("[publish] invalid file %s", path)
			continue
		}
		fsize += helper.FileSize(path)
		if fsize > config.Opts.MaxAttachSize("slack") {
			logger.Warn("total file size large than 5GB, skipped")
			continue
		}
		logger.Debug("append document: %s", path)
		reader, err := os.Open(path)
		if err != nil {
			logger.Error("open file failed: %v", err)
			continue
		}
		params := slack.FileUploadParameters{
			Filename:        path,
			Reader:          reader,
			Title:           bundle.Title,
			Channels:        []string{channel},
			ThreadTimestamp: timestamp,
		}
		file, err := client.UploadFile(params)
		if err != nil {
			logger.Error("unexpected error: %s", err)
			continue
		}
		logger.Debug("uploaded file: %#v", file)
		file, _, _, err = client.ShareFilePublicURL(file.ID)
		if err != nil {
			logger.Warn("create external link failed: %v", err)
			continue
		}
		logger.Info("slack external file permalink: %s", file.PermalinkPublic)
	}

	return nil
}
