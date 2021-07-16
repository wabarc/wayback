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
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/template/render"

	aurora "github.com/logrusorgru/aurora/v3"
	telegram "gopkg.in/tucnak/telebot.v2"
)

// Telegram handles a telegram service.
type Telegram struct {
	ctx context.Context

	bot   *telegram.Bot
	store *storage.Storage
	pool  pooling.Pool
}

// New Telegram struct.
func New(ctx context.Context, store *storage.Storage, pool pooling.Pool) *Telegram {
	if config.Opts.TelegramToken() == "" {
		logger.Fatal("missing required environment variable")
	}
	if store == nil {
		logger.Fatal("must initialize storage")
	}
	if pool == nil {
		logger.Fatal("must initialize pooling")
	}
	bot, err := telegram.NewBot(telegram.Settings{
		Token: config.Opts.TelegramToken(),
		// Verbose:   config.Opts.HasDebugMode(),
		ParseMode: telegram.ModeHTML,
		Poller:    &telegram.LongPoller{Timeout: 3 * time.Second},
	})
	if err != nil {
		logger.Fatal("create telegram bot instance failed: %v", err)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return &Telegram{
		ctx:   ctx,
		bot:   bot,
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
	logger.Info("authorized on account %s", aurora.Blue(t.bot.Me.Username))

	if channel, err := t.bot.ChatByID(config.Opts.TelegramChannel()); err == nil {
		id := strings.TrimPrefix(config.Opts.TelegramChannel(), "@")
		logger.Info("channel title: %s, channel id: %s", aurora.Blue(channel.Title), aurora.Blue(id))
	}

	go func() {
		<-t.ctx.Done()
		logger.Info("stopping receive updates...")
		t.bot.Stop()
	}()

	// Set bot commands
	t.setCommands()

	t.bot.Poller = telegram.NewMiddlewarePoller(t.bot.Poller, func(update *telegram.Update) bool {
		switch {
		case update.Callback != nil:
			logger.Debug("callback query: %#v", update.Callback)

			callback := update.Callback
			id, err := strconv.Atoi(callback.Data)
			if err != nil {
				logger.Warn("invalid playback id: %s", callback.Data)
				metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusFailure)
				return false
			}

			// Query playback callback data from database
			pb, err := t.store.Playback(id)
			if err != nil {
				logger.Error("query playback data failed: %v", err)
				metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusFailure)
				return false
			}

			data, err := base64.StdEncoding.DecodeString(pb.Source)
			if err != nil {
				logger.Error("decoding callback data failed: %v", err)
				metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusFailure)
				return false
			}

			callback.Message.Text = string(data)
			go t.process(callback.Message)
		case update.Message != nil && update.Message.FromGroup():
			logger.Debug("message: %#v", update.Message)
			if !strings.Contains(update.Message.Text, "@"+t.bot.Me.Username) {
				return false
			}
			go t.process(update.Message)
		case update.Message != nil:
			logger.Debug("message: %#v", update.Message)
			go t.process(update.Message)
		default:
			logger.Debug("update: %#v", update)
		}

		return true
	})

	logger.Info("starting receive updates...")
	t.bot.Start()

	return errors.New("done")
}

// nolint:gocyclo
func (t *Telegram) process(message *telegram.Message) (err error) {
	content := message.Text
	logger.Debug("content: %s", content)

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
		return t.playback(message)
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
		logger.Warn("archives failure, URL no found.")
		metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusRequest)
		t.reply(message, "URL no found.")
	default:
		metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusRequest)
		if message, err = t.reply(message, "Queue..."); err != nil {
			logger.Error("reply queue failed: %v", err)
			return
		}
		t.pool.Roll(func() {
			if err := t.wayback(t.ctx, message, urls); err != nil {
				logger.Error("archives failed: %v", err)
				metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusFailure)
				return
			}
			metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusSuccess)
		})
	}
	return nil
}

func (t *Telegram) wayback(ctx context.Context, message *telegram.Message, urls []string) error {
	stage, err := t.bot.Edit(message, "Archiving...")
	if err != nil {
		logger.Error("send archiving message failed: %v", err)
		return err
	}
	logger.Debug("send archiving messagee result: %v", stage)

	var bundles reduxer.Bundles
	cols, err := wayback.Wayback(ctx, &bundles, urls...)
	if err != nil {
		logger.Error("archives failure, ", err)
		return err
	}
	logger.Debug("bundles: %#v", bundles)

	replyText := render.ForReply(&render.Telegram{Cols: cols}).String()
	logger.Debug("reply text, %s", replyText)

	opts := &telegram.SendOptions{DisableWebPagePreview: true}
	if _, err := t.bot.Edit(stage, replyText, opts); err != nil {
		logger.Error("update message failed: %v", err)
		return err
	}

	ctx = context.WithValue(ctx, publish.FlagTelegram, t.bot)
	ctx = context.WithValue(ctx, publish.PubBundle, bundles)
	go publish.To(ctx, cols, publish.FlagTelegram)

	var album telegram.Album
	var fsize int64
	for _, bundle := range bundles {
		for _, path := range bundle.Paths() {
			if path == "" {
				continue
			}
			if !helper.Exists(path) {
				logger.Warn("[publish] invalid file %s", path)
				continue
			}
			fsize += helper.FileSize(path)
			if fsize > config.Opts.MaxAttachSize("telegram") {
				logger.Warn("total file size large than 50MB, skipped")
				continue
			}
			logger.Debug("append document: %s", path)
			album = append(album, &telegram.Document{
				File:     telegram.FromDisk(path),
				Caption:  bundle.Title,
				FileName: path,
			})
		}
	}
	// Send album attach files, and reply to wayback result message
	opts = &telegram.SendOptions{ReplyTo: stage, DisableNotification: true}
	if _, err := t.bot.SendAlbum(stage.Chat, album, opts); err != nil {
		logger.Error("reply failed: %v", err)
	}

	return nil
}

func (t *Telegram) playback(message *telegram.Message) error {
	metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusRequest)

	recipient, err := t.bot.ChatByID(fmt.Sprint(message.Chat.ID))
	if err != nil {
		metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusFailure)
		logger.Error("playback failed: %v", err)
		return err
	}

	urls := helper.MatchURL(message.Text)
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

	if err = t.bot.Notify(message.Sender, telegram.Typing); err != nil {
		logger.Error("send typing action failed: %v", err)
	}
	cols, _ := wayback.Playback(t.ctx, urls...)
	logger.Debug("playback collections: %#v", cols)

	// Due to Telegram restricted callback data to 1-64 bytes, it requires to store
	// playback URLs to database.
	data := []byte(strings.ReplaceAll(callbackPrefix()+message.Text, "/playback", ""))
	pb := &entity.Playback{Source: base64.StdEncoding.EncodeToString(data)}
	if err := t.store.CreatePlayback(pb); err != nil {
		logger.Error("store collections failed: %v", err)
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
	replyText := render.ForReply(&render.Telegram{Cols: cols}).String()
	if _, err := t.bot.Send(recipient, replyText, opts); err != nil {
		metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusFailure)
		logger.Error("send playback results failed: %v", err)
		return err
	}
	metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusSuccess)
	return nil
}

func (t *Telegram) reply(message *telegram.Message, text string) (*telegram.Message, error) {
	if text == "" {
		logger.Warn("text empty, skipped")
		return nil, errors.New("text empty")
	}

	opts := &telegram.SendOptions{DisableWebPagePreview: true}
	msg, err := t.bot.Reply(message, text, opts)
	if err != nil {
		logger.Error("reply failed: %v", err)
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
		logger.Error("got my failed: %v", err)
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

// nolint:stylecheck
func (t *Telegram) setCommands() (error, bool) {
	commands := t.getCommands()
	logger.Debug("got commands: %v", commands)

	if err := t.bot.SetCommands(commands); err != nil {
		logger.Error("set commands failed: %v", err)
		return err, false
	}
	logger.Debug("set commands succeed")

	return nil, true
}

func defaultCommands() []telegram.Command {
	commands := []telegram.Command{
		{
			Text:        "help",
			Description: "Show help information",
		},
		{
			Text:        "playback",
			Description: "Playback archived url",
		},
	}
	if config.Opts.EnabledMetrics() {
		commands = append(commands, telegram.Command{
			Text:        "metrics",
			Description: "Show service metrics",
		})
	}

	return commands
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
