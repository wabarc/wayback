// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"bytes"
	"text/template"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/logger"
)

type Collect struct {
	Arc string
	Dst map[string]string
}

// ToChannel for publish to message to Telegram channel,
// returns boolean as result.
func ToChannel(opts *config.Options, bot *tgbotapi.BotAPI, text string) bool {
	if bot == nil {
		var err error
		if bot, err = tgbotapi.NewBotAPI(opts.TelegramToken()); err != nil {
			logger.Error("Publish to Telegram Channel failed, %v", err)
			return false
		}
	}

	msg := tgbotapi.NewMessageToChannel("@"+opts.TelegramChannel(), text)
	msg.ParseMode = "html"
	if _, err := bot.Send(msg); err != nil {
		logger.Error("Publish message to channel failed, %v", err)
		return false
	}

	return true
}

func Render(vars []*Collect) string {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}<b>{{ $.Arc }}</b>:
{{ range $src, $dst := $.Dst -}}
• <a href="{{ $src }}">origin</a> - {{ $dst }}
{{end}}
{{end}}`

	tpl, err := template.New("message").Parse(tmpl)
	if err != nil {
		logger.Debug("Telegram: parse template failed, %v", err)
		return ""
	}

	err = tpl.Execute(&tmplBytes, vars)
	if err != nil {
		logger.Debug("Telegram: execute template failed, %v", err)
		return ""
	}

	return tmplBytes.String()
}
