// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/service/telegram"

import (
	"context"
	"fmt"
	"strings"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
)

type Telegram struct {
	bot *telegram.BotAPI
	pub *publish.Telegram
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
		pub: publish.NewTelegram(bot),
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

	go func() {
		select {
		case <-ctx.Done():
			logger.Info("[telegram] stopping receive updates...")
			t.bot.StopReceivingUpdates()
		}
	}()

	for update := range updates {
		switch {
		case update.CallbackQuery != nil:
			logger.Debug("[telegram] callback query: %#v", update.CallbackQuery)

			callback := update.CallbackQuery
			if strings.HasPrefix(callback.Data, callbackPrefix()) {
				metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusRequest)
				go t.archive(ctx, callback.Message, helper.MatchURL(callback.Data))
			}
		case update.Message != nil:
			logger.Debug("[telegram] message: %v", update.Message)

			go t.process(ctx, update)
		default:
			logger.Debug("[telegram] update: %#v", update)
		}
	}

	return errors.New("done")
}

func (t *Telegram) process(ctx context.Context, update telegram.Update) error {
	message := update.Message
	command := message.Command()
	content := message.Text
	logger.Debug("[telegram] content: %s", content)

	if message.Caption != "" {
		content = fmt.Sprintf("Text: \n%s\nCaption: \n%s", content, message.Caption)
	}
	// If the message is forwarded and contains multiple entities,
	// the update will be split into multiple parts.
	// Don't process parts of the forwarded message without text.
	if message.ForwardFromMessageID != 0 && message.Caption == "" {
		return nil
	}
	urls := helper.MatchURL(content)

	// Set command as playback if receive a playback command without URLs, and
	// required user reply a message with URLs.
	if message.ReplyToMessage != nil {
		from := message.ReplyToMessage.From
		if from.UserName == t.bot.Self.UserName {
			command = "playback"
		}
	}

	switch {
	case command == "help":
		t.reply(message, config.Opts.TelegramHelptext())
	case command == "playback", command == "search":
		return t.playback(message, urls)
	case command == "status", strings.HasPrefix(command, "metric"):
		stats := metrics.Gather.Export("wayback", "go_info")
		if config.Opts.EnabledMetrics() && stats != "" {
			return t.reply(message, stats)
		}
		return nil
	case message.IsCommand():
		commands := t.myCommands()
		if commands != "" {
			commands = fmt.Sprintf("\n\nAvailable commands:\n%s", commands)
		}
		t.reply(message, fmt.Sprintf("/%s is no specified command%s", message.Command(), commands))
	case len(urls) == 0:
		logger.Info("[telegram] archives failure, URL no found.")
		metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusRequest)
		t.reply(message, "URL no found.")
	default:
		metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusRequest)
		err := t.archive(ctx, message, urls)
		if err != nil {
			metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusFailure)
			return err
		}
		metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusSuccess)
	}
	return nil
}

func (t *Telegram) archive(ctx context.Context, message *telegram.Message, urls []string) error {
	msg := telegram.NewMessage(message.Chat.ID, "Archiving...")
	msg.ReplyToMessageID = message.MessageID
	stage, err := t.bot.Send(msg)
	if err != nil {
		logger.Error("[telegram] send archiving message failed: %v", err)
		return err
	}
	logger.Debug("[telegram] send archiving messagee result: %v", stage)
	// t.bot.Send(telegram.NewChatAction(message.Chat.ID, telegram.ChatTyping))

	col, err := wayback.Wayback(urls)
	if err != nil {
		logger.Error("[telegram] archives failure, ", err)
		return err
	}

	replyText := t.pub.Render(col)
	logger.Debug("[telegram] reply text, %s", replyText)
	updMsg := telegram.NewEditMessageText(stage.Chat.ID, stage.MessageID, replyText)
	updMsg.DisableWebPagePreview = true
	updMsg.ParseMode = "html"
	if _, err := t.bot.Send(updMsg); err != nil {
		logger.Error("[telegram] update message failed: %v", err)
		return err
	}

	ctx = context.WithValue(ctx, "telegram", t.bot)
	go publish.To(ctx, col, "telegram")

	return nil
}

func (t *Telegram) playback(message *telegram.Message, urls []string) error {
	metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusRequest)
	if len(urls) == 0 {
		msg := telegram.NewMessage(message.Chat.ID, "Please send me URLs to playback...")
		msg.ReplyToMessageID = message.MessageID
		msg.BaseChat.ReplyMarkup = telegram.ForceReply{ForceReply: true}
		if _, err := t.bot.Send(msg); err != nil {
			return err
		}
		return nil
	}

	t.bot.Send(telegram.NewChatAction(message.Chat.ID, telegram.ChatTyping))
	col, _ := wayback.Playback(urls)

	msg := telegram.NewMessage(message.Chat.ID, t.pub.Render(col))
	msg.ReplyToMessageID = message.MessageID
	// Attach a button below the message to send a wayback request quickly
	msg.BaseChat.ReplyMarkup = telegram.NewInlineKeyboardMarkup(telegram.NewInlineKeyboardRow(
		telegram.NewInlineKeyboardButtonData("wayback", callbackPrefix()+message.Text),
	))
	msg.DisableWebPagePreview = true
	msg.ParseMode = "html"
	if _, err := t.bot.Send(msg); err != nil {
		metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusFailure)
		return err
	}
	metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusSuccess)
	return nil
}

func (t *Telegram) reply(message *telegram.Message, text string) error {
	msg := telegram.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID
	if _, err := t.bot.Send(msg); err != nil {
		logger.Error("[telegram] reply failed: %v", err)
		return err
	}
	return nil
}

func (t *Telegram) myCommands() string {
	commands, err := t.bot.GetMyCommands()
	if err != nil {
		return ""
	}

	var list string
	for _, command := range commands {
		list = fmt.Sprintf("/%s - %s\n", command.Command, command.Description)
	}

	return list
}

func callbackPrefix() string {
	return ":wayback "
}
