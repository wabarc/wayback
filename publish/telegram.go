// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"

	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"
	telegram "gopkg.in/telebot.v3"
)

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
// A context should contain a `reduxer.Bundle` via `publish.PubBundle` constant.
func (t *telegramBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Warn("collects empty")
		return
	}

	var bnd = bundle(ctx, cols)
	var txt = render.ForPublish(&render.Telegram{Cols: cols, Data: bnd}).String()
	if t.toChannel(bnd, txt) {
		metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusFailure)
	return
}

// toChannel for publish to message to Telegram channel,
// returns boolean as result.
func (t *telegramBot) toChannel(bundle *reduxer.Bundle, text string) (ok bool) {
	if text == "" {
		logger.Warn("post to message to channel failed, text empty")
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

	stage, err := t.bot.Send(chat, text)
	if err != nil {
		logger.Error("post message to channel failed, %v", err)
		return ok
	}

	if bundle == nil {
		logger.Warn("bundle empty")
		return true
	}

	album := UploadToTelegram(bundle)
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

// UploadToTelegram composes files into an album by the given bundle.
func UploadToTelegram(bundle *reduxer.Bundle) telegram.Album {
	// Attach image and pdf files
	var album telegram.Album
	var fsize int64
	for _, asset := range bundle.Asset() {
		if asset.Local == "" {
			continue
		}
		if !helper.Exists(asset.Local) {
			logger.Warn("invalid file %s", asset.Local)
			continue
		}
		fsize += helper.FileSize(asset.Local)
		if fsize > config.Opts.MaxAttachSize("telegram") {
			logger.Warn("total file size large than 50MB, skipped")
			continue
		}
		logger.Debug("append document: %s", asset.Local)
		album = append(album, &telegram.Document{
			File:     telegram.FromDisk(asset.Local),
			Caption:  bundle.Title,
			FileName: asset.Local,
		})
	}
	return album
}
