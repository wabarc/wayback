// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/service/telegram"

import (
	"context"
	"fmt"
	"sync"

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

// New telegram struct.
func New(opts *config.Options) *telegram {
	return &telegram{
		opts: opts,
	}
}

// Serve loop request message from the Telegram api server.
// Serve always returns a nil error.
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
			col, msgID, err := archive(t, update.Message.MessageID, urls)
			if err != nil {
				logger.Error("Telegram: archiving failed, ", err)
				return err
			}

			replyText := publish.Render(col)
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

func archive(t *telegram, msgid int, urls []string) (col []*publish.Collect, id int, err error) {
	logger.Debug("Telegram: archives start...")

	wg := sync.WaitGroup{}
	var wbrc wayback.Broker = &wayback.Handle{URLs: urls, Opts: t.opts}
	for slot, arc := range t.opts.Slots() {
		if !arc {
			continue
		}
		wg.Add(1)
		go func(slot string) {
			defer wg.Done()
			c := &publish.Collect{}
			switch slot {
			case config.SLOT_IA:
				logger.Debug("Telegram: archiving slot: %s", slot)
				c.Arc = fmt.Sprintf("<a href='https://web.archive.org/'>%s</a>", config.SlotName(slot))
				c.Dst = wbrc.IA()
			case config.SLOT_IS:
				logger.Debug("Telegram: archiving slot: %s", slot)
				c.Arc = fmt.Sprintf("<a href='https://archive.today/'>%s</a>", config.SlotName(slot))
				c.Dst = wbrc.IS()
			case config.SLOT_IP:
				logger.Debug("Telegram: archiving slot: %s", slot)
				c.Arc = fmt.Sprintf("<a href='https://ipfs.github.io/public-gateway-checker/'>%s</a>", config.SlotName(slot))
				c.Dst = wbrc.IP()
			}
			col = append(col, c)
		}(slot)
	}
	wg.Wait()

	return col, msgid, nil
}
