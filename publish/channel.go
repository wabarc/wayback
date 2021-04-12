// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"bytes"
	"text/template"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/logger"
)

// ToChannel for publish to message to Telegram channel,
// returns boolean as result.
func ToChannel(bot *telegram.BotAPI, text string) bool {
	if bot == nil {
		var err error
		if bot, err = telegram.NewBotAPI(config.Opts.TelegramToken()); err != nil {
			logger.Error("[publish] post to Telegram Channel failed, %v", err)
			return false
		}
	}

	msg := telegram.NewMessageToChannel("@"+config.Opts.TelegramChannel(), text)
	msg.ParseMode = "html"
	if _, err := bot.Send(msg); err != nil {
		logger.Error("[publish] post message to channel failed, %v", err)
		return false
	}

	return true
}

func Render(vars []*wayback.Collect) string {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}<b><a href='{{ $.Ext }}'>{{ $.Arc }}</a></b>:
{{ range $src, $dst := $.Dst -}}
• <a href="{{ $src }}">origin</a> - {{ $dst }}
{{end}}
{{end}}`

	tpl, err := template.New("message").Parse(tmpl)
	if err != nil {
		logger.Debug("[publish] parse Telegram template failed, %v", err)
		return ""
	}

	err = tpl.Execute(&tmplBytes, vars)
	if err != nil {
		logger.Debug("[publish] execute Telegram template failed, %v", err)
		return ""
	}

	return tmplBytes.String()
}
