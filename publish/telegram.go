// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/reduxer"
	telegram "gopkg.in/tucnak/telebot.v2"
)

type Telegram struct {
	bot *telegram.Bot
}

// NewTelegram returns Telegram bot client
func NewTelegram(bot *telegram.Bot) *Telegram {
	if !config.Opts.PublishToChannel() {
		logger.Error("Missing required environment variable, abort.")
		return new(Telegram)
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

	return &Telegram{bot: bot}
}

// ToChannel for publish to message to Telegram channel,
// returns boolean as result.
func (t *Telegram) ToChannel(ctx context.Context, text string) (ok bool) {
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
	if head := title(ctx, text); head != "" {
		text = "<b>" + head + "</b>\n\n" + text
	}
	stage, err := t.bot.Send(chat, text)
	if err != nil {
		logger.Error("[publish] post message to channel failed, %v", err)
		return ok
	}

	var bundles []reduxer.Bundle
	if bundles, ok = ctx.Value(PubBundle).([]reduxer.Bundle); !ok {
		logger.Debug("[publish] bundles empty")
		return true
	}

	// Attach image and pdf files
	var album telegram.Album
	for _, bundle := range bundles {
		paths := []string{
			bundle.Path.Img,
			bundle.Path.PDF,
		}
		for _, path := range paths {
			if path == "" {
				logger.Info("[publish] invalid file path: %s", path)
				continue
			}
			logger.Debug("[publish] append document: %s", path)
			album = append(album, &telegram.Document{
				File:     telegram.FromDisk(path),
				Caption:  bundle.Title,
				FileName: path,
			})
		}
	}
	// Send album attach files, and reply to wayback result message
	opts := &telegram.SendOptions{ReplyTo: stage, DisableNotification: true}
	if _, err := t.bot.SendAlbum(stage.Chat, album, opts); err != nil {
		logger.Error("[publish] reply failed: %v", err)
	}

	return true
}

func (t *Telegram) Render(vars []wayback.Collect) string {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}<b><a href='{{ $.Ext }}'>{{ $.Arc }}</a></b>:
{{ range $src, $dst := $.Dst -}}
• <a href="{{ $src | revert }}">source</a> - {{ if $dst | isURL }}<a href="{{ $dst }}">{{ $dst }}</a>{{ else }}{{ $dst }}{{ end }}
{{end}}
{{end}}`

	tpl, err := template.New("message").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Debug("[publish] parse Telegram template failed, %v", err)
		return ""
	}

	if err = tpl.Execute(&tmplBytes, vars); err != nil {
		logger.Debug("[publish] execute Telegram template failed, %v", err)
		return ""
	}

	return strings.TrimSpace(tmplBytes.String()) + "\n\n#wayback #存档"
}
