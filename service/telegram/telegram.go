// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/service/telegram"

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/entity"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/storage"
	telegram "gopkg.in/tucnak/telebot.v2"
)

// Telegram handles a telegram service.
type Telegram struct {
	ctx context.Context

	bot   *telegram.Bot
	pub   *publish.Telegram
	store *storage.Storage
	pool  pooling.Pool
}

// New Telegram struct.
func New(ctx context.Context, store *storage.Storage, pool pooling.Pool) *Telegram {
	if config.Opts.TelegramToken() == "" {
		logger.Fatal("[telegram] missing required environment variable")
	}
	if store == nil {
		logger.Fatal("[telegram] must initialize storage")
	}
	if pool == nil {
		logger.Fatal("[telegram] must initialize pooling")
	}
	bot, err := telegram.NewBot(telegram.Settings{
		Token: config.Opts.TelegramToken(),
		// Verbose:   config.Opts.HasDebugMode(),
		ParseMode: telegram.ModeHTML,
		Poller:    &telegram.LongPoller{Timeout: 3 * time.Second},
	})
	if err != nil {
		logger.Fatal("[telegram] create telegram bot instance failed: %v", err)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return &Telegram{
		ctx:   ctx,
		bot:   bot,
		pub:   publish.NewTelegram(bot),
		store: store,
		pool:  pool,
	}
}

// Serve loop request message from the Telegram api server.
// Serve always returns an error.
func (t *Telegram) Serve() (err error) {
	if t.bot == nil {
		return errors.New("Initialize telegram failed, error: %v", err)
	}
	logger.Info("[telegram] authorized on account %s", t.bot.Me.Username)

	go func() {
		<-t.ctx.Done()
		logger.Info("[telegram] stopping receive updates...")
		t.bot.Stop()
	}()

	// Set bot commands
	t.setCommands()

	t.bot.Poller = telegram.NewMiddlewarePoller(t.bot.Poller, func(update *telegram.Update) bool {
		switch {
		case update.Callback != nil:
			logger.Debug("[telegram] callback query: %#v", update.Callback)

			callback := update.Callback
			id, err := strconv.Atoi(callback.Data)
			if err != nil {
				logger.Error("[telegram] invalid playback id: %s", callback.Data)
				metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusFailure)
				return false
			}

			// Query playback callback data from database
			pb, err := t.store.Playback(id)
			if err != nil {
				logger.Error("[telegram] query playback data failed: %v", err)
				metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusFailure)
				return false
			}

			data, err := base64.StdEncoding.DecodeString(pb.Source)
			if err != nil {
				logger.Error("[telegram] decoding callback data failed: %v", err)
				metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusFailure)
				return false
			}

			callback.Message.Text = string(data)
			go t.process(callback.Message)
		case update.Message != nil && update.Message.FromGroup():
			logger.Debug("[telegram] message: %#v", update.Message)
			if !strings.Contains(update.Message.Text, "@"+t.bot.Me.Username) {
				return false
			}
			go t.process(update.Message)
		case update.Message != nil:
			logger.Debug("[telegram] message: %#v", update.Message)
			go t.process(update.Message)
		default:
			logger.Debug("[telegram] update: %#v", update)
		}

		return true
	})

	logger.Info("[telegram] starting receive updates...")
	t.bot.Start()

	return errors.New("done")
}

func (t *Telegram) process(message *telegram.Message) (err error) {
	content := message.Text
	logger.Debug("[telegram] content: %s", content)

	if message.Caption != "" {
		content = fmt.Sprintf("Text: \n%s\nCaption: \n%s", content, message.Caption)
	}
	// If the message is forwarded and contains multiple entities,
	// the update will be split into multiple parts.
	// Don't process parts of the forwarded message without text.
	// if message.IsForwarded() && message.Caption == "" {
	if message.IsForwarded() && content == "" {
		return nil
	}
	urls := helper.MatchURLFallback(content)

	// Set command as playback if receive a playback command without URLs, and
	// required user reply a message with URLs.
	if message.IsReply() {
		if message.ReplyTo.Sender.Username == t.bot.Me.Username {
			content = "/playback" + content
		}
	}

	command := command(content)
	switch {
	case command == "help", command == "start":
		t.reply(message, config.Opts.TelegramHelptext())
	case command == "playback":
		return t.playback(message, urls)
	case command == "metrics":
		stats := metrics.Gather.Export("wayback")
		if config.Opts.EnabledMetrics() && stats != "" {
			if _, err = t.reply(message, stats); err != nil {
				return err
			}
		}
		return nil
	case command != "":
		fallback := t.commandFallback()
		if fallback != "" {
			fallback = fmt.Sprintf("\n\nAvailable commands:\n%s", fallback)
		}
		t.reply(message, fmt.Sprintf("/%s is an illegal command%s", command, fallback))
	case len(urls) == 0:
		logger.Info("[telegram] archives failure, URL no found.")
		metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusRequest)
		t.reply(message, "URL no found.")
	default:
		metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusRequest)
		if message, err = t.reply(message, "Queue..."); err != nil {
			logger.Error("[telegram] reply queue failed: %v", err)
			return
		}
		t.pool.Roll(func() {
			if err := t.archive(t.ctx, message, urls); err != nil {
				logger.Error("[telegram] archives failed: %v", err)
				metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusFailure)
				return
			}
			metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusSuccess)
		})
	}
	return nil
}

func (t *Telegram) archive(ctx context.Context, message *telegram.Message, urls []string) error {
	stage, err := t.bot.Edit(message, "Archiving...")
	if err != nil {
		logger.Error("[telegram] send archiving message failed: %v", err)
		return err
	}
	logger.Debug("[telegram] send archiving messagee result: %v", stage)

	col, err := wayback.Wayback(urls)
	if err != nil {
		logger.Error("[telegram] archives failure, ", err)
		return err
	}

	replyText := t.pub.Render(col)
	logger.Debug("[telegram] reply text, %s", replyText)

	opts := &telegram.SendOptions{DisableWebPagePreview: true}
	if _, err := t.bot.Edit(stage, replyText, opts); err != nil {
		logger.Error("[telegram] update message failed: %v", err)
		return err
	}

	ctx = context.WithValue(ctx, publish.FlagTelegram, t.bot)
	go publish.To(ctx, col, publish.FlagTelegram)

	return nil
}

func (t *Telegram) playback(message *telegram.Message, urls []string) error {
	metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusRequest)

	recipient, err := t.bot.ChatByID(fmt.Sprint(message.Chat.ID))
	if err != nil {
		metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusFailure)
		logger.Error("[telegram] playback failed: %v", err)
		return err
	}

	if len(urls) == 0 {
		opts := &telegram.SendOptions{
			ReplyTo:               message,
			DisableWebPagePreview: true,
			ReplyMarkup: &telegram.ReplyMarkup{
				ForceReply: true,
			},
		}
		_, err = t.bot.Send(recipient, "Please send me URLs to playback...", opts)
		if err != nil {
			return err
		}
		return nil
	}

	if err = t.bot.Notify(message.Sender, telegram.ChatAction(telegram.Typing)); err != nil {
		logger.Error("[telegram] send typing action failed: %v", err)
	}
	col, _ := wayback.Playback(urls)
	logger.Debug("[telegram] playback collections: %#v", col)

	// Due to Telegram restricted callback data to 1-64 bytes, it requires to store
	// playback URLs to database.
	data := []byte(strings.ReplaceAll(callbackPrefix()+message.Text, "/playback", ""))
	pb := &entity.Playback{Source: base64.StdEncoding.EncodeToString(data)}
	if err := t.store.CreatePlayback(pb); err != nil {
		logger.Error("[telegram] store collections failed: %v", err)
		return err
	}

	opts := &telegram.SendOptions{
		ReplyTo:               message,
		DisableWebPagePreview: true,
		ReplyMarkup: &telegram.ReplyMarkup{
			InlineKeyboard: [][]telegram.InlineButton{
				{{
					Text: "wayback",
					Data: strconv.Itoa(pb.ID),
				}},
			},
		},
	}
	if _, err := t.bot.Send(recipient, t.pub.Render(col), opts); err != nil {
		metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusFailure)
		logger.Error("[telegram] send playback results failed: %v", err)
		return err
	}
	metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusSuccess)
	return nil
}

func (t *Telegram) reply(message *telegram.Message, text string) (*telegram.Message, error) {
	opts := &telegram.SendOptions{DisableWebPagePreview: true}
	msg, err := t.bot.Reply(message, text, opts)
	if err != nil {
		logger.Error("[telegram] reply failed: %v", err)
		return nil, err
	}
	return msg, nil
}

func (t *Telegram) commandFallback() string {
	commands := t.getCommands()

	var list string
	for _, command := range commands {
		list += fmt.Sprintf("/%s - %s\n", command.Text, command.Description)
	}

	return list
}

func (t *Telegram) getCommands() []telegram.Command {
	commands, err := t.bot.GetCommands()
	if err != nil {
		logger.Error("[telegram] got my failed: %v", err)
	}

	var maps = make(map[string]bool, len(commands))
	for _, command := range commands {
		maps[command.Text] = true
	}

	for _, command := range defaultCommands() {
		if maps[command.Text] {
			continue
		}
		commands = append(commands, command)
	}

	return commands
}

func (t *Telegram) setCommands() (error, bool) {
	commands := t.getCommands()
	logger.Debug("[telegram] got commands: %v", commands)

	if err := t.bot.SetCommands(commands); err != nil {
		logger.Error("[telegram] set commands failed: %v", err)
		return err, false
	}
	logger.Debug("[telegram] set commands succeed")

	return nil, true
}

func defaultCommands() []telegram.Command {
	return []telegram.Command{
		{
			Text:        "help",
			Description: "Show help information",
		},
		{
			Text:        "metrics",
			Description: "Show service metrics",
		},
		{
			Text:        "playback",
			Description: "Playback archived url",
		},
	}
}

func callbackPrefix() string {
	return ":wayback "
}

func command(message string) string {
	matchCmd := func(str string) string {
		re := regexp.MustCompile(`(?m)^\/\w+`)
		for _, match := range re.FindAllString(str, -1) {
			return strings.TrimLeft(match, "/")
		}
		return ""
	}

	switch {
	case strings.HasPrefix(message, "/help"), strings.HasPrefix(message, "/start"):
		return "help"
	case strings.HasPrefix(message, "/playback"):
		return "playback"
	case strings.HasPrefix(message, "/metrics"):
		return "metrics"
	default:
		return matchCmd(message)
	}
}
