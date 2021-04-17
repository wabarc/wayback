// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/service/telegram"

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/publish"
)

type Telegram struct {
	bot *telegram.BotAPI
}

// New Telegram struct.
func New() *Telegram {
	if config.Opts.TelegramToken() == "" {
		logger.Fatal("[telegram] missing required environment variable")
	}
	bot, err := telegram.NewBotAPI(config.Opts.TelegramToken())
	if err != nil {
		logger.Fatal("[telegram] create telegram bot instance failed: %v", err)
	}

	return &Telegram{
		bot: bot,
	}
}

// Serve loop request message from the Telegram api server.
// Serve always returns an error.
func (t *Telegram) Serve(ctx context.Context) (err error) {
	if t.bot == nil {
		return errors.New("Initialize telegram failed, error: %v", err)
	}
	logger.Info("[telegram] authorized on account %s", t.bot.Self.UserName)

	cfg := telegram.NewUpdate(0)
	cfg.Timeout = 60
	updates := t.bot.GetUpdatesChan(cfg)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Info("[telegram] stopping receive updates...")
		t.bot.StopReceivingUpdates()
	}()

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		logger.Debug("[telegram] message: %v", update.Message)

		go t.process(ctx, update)
	}

	return errors.New("done")
}

func (t *Telegram) process(ctx context.Context, update telegram.Update) error {
	message := update.Message
	text := message.Text
	logger.Debug("[telegram] message: %s", text)

	if message.Caption != "" {
		text = fmt.Sprintf("Text: \n%s\nCaption: \n%s", text, message.Caption)
	}
	urls := helper.MatchURL(text)
	tel := publish.NewTelegram(t.bot)
	switch {
	case message.Command() == "help":
		msg := telegram.NewMessage(message.Chat.ID, config.Opts.TelegramHelptext())
		msg.ReplyToMessageID = message.MessageID
		t.bot.Send(msg)
		return nil
	case message.Command() == "playback", message.Command() == "search":
		col, _ := wayback.Playback(urls)
		msg := telegram.NewMessage(message.Chat.ID, tel.Render(col))
		msg.ReplyToMessageID = message.MessageID
		msg.ParseMode = "html"
		t.bot.Send(msg)
		return nil
	case message.IsCommand():
		msg := telegram.NewMessage(message.Chat.ID, fmt.Sprintf("/%s no specified", message.Command()))
		msg.ReplyToMessageID = message.MessageID
		t.bot.Send(msg)
		return nil
	case len(urls) == 0:
		logger.Info("[telegram] archives failure, URL no found.")
		msg := telegram.NewMessage(message.Chat.ID, "URL no found.")
		msg.ReplyToMessageID = message.MessageID
		t.bot.Send(msg)
		return nil
	}

	col, err := t.archive(urls)
	if err != nil {
		logger.Error("[telegram] archives failure, ", err)
		return err
	}

	replyText := tel.Render(col)
	logger.Debug("[telegram] reply text, %s", replyText)
	msg := telegram.NewMessage(message.Chat.ID, replyText)
	msg.ReplyToMessageID = message.MessageID
	msg.ParseMode = "html"

	t.bot.Send(msg)

	ctx = context.WithValue(ctx, "telegram", t.bot)
	go publish.To(ctx, col, "telegram")

	return nil
}

func (t *Telegram) archive(urls []string) (col []*wayback.Collect, err error) {
	logger.Debug("[telegram] archives start...")

	wg := sync.WaitGroup{}
	var wbrc wayback.Broker = &wayback.Handle{URLs: urls}
	for slot, arc := range config.Opts.Slots() {
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
