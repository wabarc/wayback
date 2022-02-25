// Copyright 2020 Wayback Archiver. All rights reserved.
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

	telegram "gopkg.in/telebot.v3"
)

var _ Publisher = (*telegramBot)(nil)

type telegramBot struct {
	bot *telegram.Bot
}

// NewTelegram returns Telegram bot client
func NewTelegram(bot *telegram.Bot) *telegramBot {
	if !config.Opts.PublishToChannel() {
		logger.Error("Missing required environment variable, abort.")
		return new(telegramBot)
	}

	if bot == nil {
		var err error
		if bot, err = telegram.NewBot(telegram.Settings{
			Token:     config.Opts.TelegramToken(),
			Verbose:   config.Opts.HasDebugMode(),
			ParseMode: telegram.ModeHTML,
		}); err != nil {
			logger.Error("create telegram bot instance failed: %v", err)
		}
	}

	return &telegramBot{bot: bot}
}

// Publish publish text to the Telegram channel of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (t *telegramBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusRequest)

	if len(cols) == 0 {
		return errors.New("publish to telegram: collects empty")
	}

	rdx, art, err := extract(ctx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	var head = render.Title(cols, rdx)
	var body = render.ForPublish(&render.Telegram{Cols: cols, Data: rdx}).String()
	if t.toChannel(art, head, body) {
		metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusSuccess)
		return nil
	}
	metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusFailure)
	return errors.New("publish to telegram failed")
}

// toChannel for publish to message to Telegram channel,
// returns boolean as result.
func (t *telegramBot) toChannel(art reduxer.Artifact, head, body string) (ok bool) {
	if body == "" {
		logger.Warn("post to message to channel failed, body empty")
		return ok
	}
	if t.bot == nil {
		var err error
		if t.bot, err = telegram.NewBot(telegram.Settings{
			Token:     config.Opts.TelegramToken(),
			Verbose:   config.Opts.HasDebugMode(),
			ParseMode: telegram.ModeHTML,
		}); err != nil {
			logger.Error("post to channel failed, %v", err)
			return ok
		}
	}

	chat, err := t.bot.ChatByUsername(config.Opts.TelegramChannel())
	if err != nil {
		logger.Error("open a chat failed: %v", err)
		return ok
	}

	stage, err := t.bot.Send(chat, body)
	if err != nil {
		logger.Error("post message to channel failed, %v", err)
		return ok
	}

	album := service.UploadToTelegram(art, head)
	if len(album) == 0 {
		return true
	}
	// Send album attach files, and reply to wayback result message
	opts := &telegram.SendOptions{ReplyTo: stage, DisableNotification: true}
	if _, err := t.bot.SendAlbum(stage.Chat, album, opts); err != nil {
		logger.Error("reply failed: %v", err)
	}

	return true
}
