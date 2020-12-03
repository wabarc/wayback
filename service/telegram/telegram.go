// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/sevice/telegram"

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"text/template"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/logger"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/utils"
	"golang.org/x/sync/errgroup"
)

type telegram struct {
	opts *config.Options
}

func New(opts *config.Options) *telegram {
	return &telegram{
		opts: opts,
	}
}

func (t *telegram) Serve(_ context.Context) error {
	bot, err := tgbotapi.NewBotAPI(t.opts.TelegramToken())
	if err != nil {
		return errors.New("Initialize telegram failed, error: %v", err)
	}

	logger.Info("Telegram: authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 600

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return errors.New("Get telegram message channel failed, error: %v", err)
	}

	g, _ := errgroup.WithContext(context.Background())
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		message := update.Message
		text := message.Text
		logger.Debug("Telegram: message: %s", text)

		urls := utils.MatchURL(text)
		switch {
		case message.IsCommand():
			continue
		case len(urls) == 0:
			logger.Info("Telegram: archives failure, URL no found.")
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "URL no found.")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			continue
		}

		g.Go(func() error {
			col := &collect{}
			msgID, err := col.archive(t, update.Message.MessageID, urls)
			if err != nil {
				logger.Error("Telegram: archiving failed, ", err)
				return err
			}

			replyText := render(col)
			logger.Debug("Telegram: reply text, %s", replyText)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyText)
			msg.ReplyToMessageID = msgID
			msg.ParseMode = "html"

			bot.Send(msg)

			if t.opts.TelegramChannel() != "" {
				logger.Debug("Telegram: publishing to channel...")
				publish.ToChannel(t.opts, bot, replyText)
			}
			return nil
		})
	}

	return nil
}

type collect struct {
	Arc []string
	Dst []map[string]string
}

func (c *collect) archive(t *telegram, msgid int, urls []string) (int, error) {
	logger.Debug("Telegram: archives start...")
	p := *c

	wg := sync.WaitGroup{}
	var wbrc wayback.Broker = &wayback.Handle{URLs: urls, Opts: t.opts}
	for slot, arc := range t.opts.Slots() {
		if !arc {
			continue
		}
		wg.Add(1)
		go func(slot string) {
			defer wg.Done()
			switch slot {
			case config.SLOT_IA:
				logger.Debug("Telegram: archiving slot: %s", slot)
				p.Dst = append(p.Dst, wbrc.IA())
				p.Arc = append(p.Arc, fmt.Sprintf("<a href='https://web.archive.org/'>%s</a>", config.SlotName(slot)))
			case config.SLOT_IS:
				logger.Debug("Telegram: archiving slot: %s", slot)
				p.Dst = append(p.Dst, wbrc.IS())
				p.Arc = append(p.Arc, fmt.Sprintf("<a href='https://archive.today/'>%s</a>", config.SlotName(slot)))
			case config.SLOT_IP:
				logger.Debug("Telegram: archiving slot: %s", slot)
				p.Dst = append(p.Dst, wbrc.IP())
				p.Arc = append(p.Arc, fmt.Sprintf("<a href='https://ipfs.github.io/public-gateway-checker/'>%s</a>", config.SlotName(slot)))
			}
		}(slot)
	}
	wg.Wait()
	*c = p

	return msgid, nil
}

func render(vars *collect) string {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $i, $name := .Arc}}<b>{{ $name }}</b>:
{{ range $url := index $.Dst $i -}}
• {{ $url }}
{{end}}
{{end}}`

	tpl, err := template.New("message").Parse(tmpl)
	if err != nil {
		return ""
	}

	err = tpl.Execute(&tmplBytes, vars)
	if err != nil {
		return ""
	}

	return tmplBytes.String()
}
