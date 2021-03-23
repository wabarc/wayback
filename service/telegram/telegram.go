// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/service/telegram"

import (
	"context"
	"sync"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/logger"
	"github.com/wabarc/wayback/publish"
)

type Telegram struct {
	opts *config.Options

	bot *tgbotapi.BotAPI
}

// New Telegram struct.
func New(opts *config.Options) *Telegram {
	return &Telegram{
		opts: opts,
	}
}

// Serve loop request message from the Telegram api server.
// Serve always returns an error.
func (t *Telegram) Serve(ctx context.Context) (err error) {
	if t.bot, err = tgbotapi.NewBotAPI(t.opts.TelegramToken()); err != nil {
		return errors.New("Initialize telegram failed, error: %v", err)
	}

	logger.Info("[telegram] authorized on account %s", t.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := t.bot.GetUpdatesChan(u)
	if err != nil {
		return errors.New("Get telegram message channel failed, error: %v", err)
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		go t.process(ctx, update)
	}

	return errors.New("done")
}

func (t *Telegram) process(ctx context.Context, update tgbotapi.Update) {
	bot := t.bot
	message := update.Message
	text := message.Text
	logger.Debug("[telegram] message: %s", text)

	urls := helper.MatchURL(text)
	switch {
	case message.IsCommand():
		return
	case len(urls) == 0:
		logger.Info("[telegram] archives failure, URL no found.")
		msg := tgbotapi.NewMessage(message.Chat.ID, "URL no found.")
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
		return
	}

	col, err := t.archive(urls)
	if err != nil {
		logger.Error("[telegram] archives failure, ", err)
		return
	}

	replyText := publish.Render(col)
	logger.Debug("[telegram] reply text, %s", replyText)
	msg := tgbotapi.NewMessage(message.Chat.ID, replyText)
	msg.ReplyToMessageID = message.MessageID
	msg.ParseMode = "html"

	bot.Send(msg)

	go publish.To(ctx, t.opts, col, "telegram")
}

func (t *Telegram) archive(urls []string) (col []*wayback.Collect, err error) {
	logger.Debug("[telegram] archives start...")

	wg := sync.WaitGroup{}
	var wbrc wayback.Broker = &wayback.Handle{URLs: urls, Opts: t.opts}
	for slot, arc := range t.opts.Slots() {
		if !arc {
			continue
		}
		wg.Add(1)
		go func(slot string) {
			defer wg.Done()
			c := &wayback.Collect{}
			logger.Debug("[telegram] archiving slot: %s", slot)
			switch slot {
			case config.SLOT_IA:
				c.Dst = wbrc.IA()
			case config.SLOT_IS:
				c.Dst = wbrc.IS()
			case config.SLOT_IP:
				c.Dst = wbrc.IP()
			case config.SLOT_PH:
				c.Dst = wbrc.PH()
			}
			c.Arc = config.SlotName(slot)
			c.Ext = config.SlotExtra(slot)
			col = append(col, c)
		}(slot)
	}
	wg.Wait()

	return col, nil
}
