// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"strings"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"
	telegram "gopkg.in/tucnak/telebot.v2"
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
			logger.Error("[telegram] create telegram bot instance failed: %v", err)
		}
	}

	return &telegramBot{bot: bot}
}

func (t *telegramBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Debug("[publish] collects empty")
		return
	}

	var bnd = bundle(ctx, cols)
	var txt = render.ForPublish(&render.Telegram{Cols: cols}).String()
	if t.toChannel(ctx, &bnd, txt) {
		metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusFailure)
	return
}

// toChannel for publish to message to Telegram channel,
// returns boolean as result.
func (t *telegramBot) toChannel(ctx context.Context, bundle *reduxer.Bundle, text string) (ok bool) {
	if text == "" {
		logger.Error("[publish] post to message to channel failed, text empty")
		return ok
	}
	if t.bot == nil {
		var err error
		if t.bot, err = telegram.NewBot(telegram.Settings{
			Token:     config.Opts.TelegramToken(),
			Verbose:   config.Opts.HasDebugMode(),
			ParseMode: telegram.ModeHTML,
		}); err != nil {
			logger.Error("[publish] post to channel failed, %v", err)
			return ok
		}
	}

	chat, err := t.bot.ChatByID(config.Opts.TelegramChannel())
	if err != nil {
		logger.Error("[publish] open a chat failed: %v", err)
		return ok
	}

	var b strings.Builder
	if head := title(ctx, bundle); head != "" {
		b.WriteString("<b>")
		b.WriteString(head)
		b.WriteString("</b>\n\n")
	}
	if dgst := digest(ctx, bundle); dgst != "" {
		b.WriteString(dgst)
		b.WriteString("\n\n")
	}
	b.WriteString(text)

	stage, err := t.bot.Send(chat, b.String())
	if err != nil {
		logger.Error("[publish] post message to channel failed, %v", err)
		return ok
	}

	if bundle == nil {
		logger.Debug("[publish] bundle empty")
		return true
	}

	// Attach image and pdf files
	var album telegram.Album
	paths := []string{
		bundle.Path.Img,
		bundle.Path.PDF,
		bundle.Path.WARC,
	}

	for _, path := range paths {
		if path == "" {
			logger.Info("[publish] invalid file path, skipped")
			continue
		}
		logger.Debug("[publish] append document: %s", path)
		album = append(album, &telegram.Document{
			File:     telegram.FromDisk(path),
			Caption:  bundle.Title,
			FileName: path,
		})
	}
	if len(album) == 0 {
		return true
	}
	// Send album attach files, and reply to wayback result message
	opts := &telegram.SendOptions{ReplyTo: stage, DisableNotification: true}
	if _, err := t.bot.SendAlbum(stage.Chat, album, opts); err != nil {
		logger.Error("[publish] reply failed: %v", err)
	}

	return true
}
