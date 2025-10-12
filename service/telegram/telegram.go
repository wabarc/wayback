// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/service/telegram"

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
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
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/template/render"

	telegram "gopkg.in/telebot.v3"
)

// Interface guard
var _ service.Servicer = (*Telegram)(nil)

// ErrServiceClosed is returned by the Service's Serve method after a call to Shutdown.
var ErrServiceClosed = errors.New("telegram: Service closed")

var (
	pollTick = 3 * time.Second
	space    = ` `
)

// Telegram represents a Telegram service in the application.
type Telegram struct {
	ctx context.Context

	bot   *telegram.Bot
	store *storage.Storage
	opts  *config.Options
	pool  *pooling.Pool
	pub   *publish.Publish
}

// New Telegram struct.
func New(ctx context.Context, opts service.Options) (*Telegram, error) {
	if !opts.Config.TelegramEnabled() {
		return nil, errors.New("missing required environment variable, skipped")
	}
	bot, err := telegram.NewBot(telegram.Settings{
		Token: opts.Config.TelegramToken(),
		// Verbose:   opts.Config.HasDebugMode(),
		ParseMode: telegram.ModeHTML,
		Poller:    &telegram.LongPoller{Timeout: pollTick},
		OnError: func(err error, _ telegram.Context) {
			if err != nil {
				logger.Warn(err.Error())
			}
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "create telegram bot instance failed")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return &Telegram{
		ctx:   ctx,
		bot:   bot,
		store: opts.Storage,
		opts:  opts.Config,
		pool:  opts.Pool,
		pub:   opts.Publish,
	}, nil
}

// Serve loop request message from the Telegram api server.
// Serve always returns an error.
func (t *Telegram) Serve() (err error) {
	if t.bot == nil {
		return errors.New("Initialize telegram failed, error: %v", err)
	}
	logger.Info("authorized on account %s", color.Blue.Sprint(t.bot.Me.Username))

	if channel, err := t.bot.ChatByUsername(t.opts.TelegramChannel()); err == nil {
		id := strconv.FormatInt(channel.ID, 10)
		logger.Info("channel title: %s, channel id: %s", color.Blue.Sprint(channel.Title), color.Blue.Sprint(id))
	}

	// Set bot commands
	// nolint:errcheck
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
			x := strconv.Itoa(id)
			u, err := strconv.ParseUint(x, 10, 64)
			if err != nil {
				logger.Error("parse uint failed: %v", err)
				return false
			}
			pb, err := t.store.Playback(u)
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

			callback.Message.Text = helper.Byte2String(data)
			go t.process(callback.Message) // nolint:errcheck
		case update.Message != nil && update.Message.FromGroup():
			transform(update.Message)
			logger.Debug("message: %#v", update.Message)
			// Reply message and mention bot on the group
			if update.Message.ReplyTo != nil {
				update.Message.Text += update.Message.ReplyTo.Text
			}
			if !strings.Contains(update.Message.Text, "@"+t.bot.Me.Username) {
				return false
			}
			go t.process(update.Message) // nolint:errcheck
		case update.Message != nil:
			// Ignore auto-delete timer message
			if update.Message.AutoDeleteTimer != nil {
				return false
			}

			transform(update.Message)
			logger.Debug("message: %#v", update.Message)
			go t.process(update.Message) // nolint:errcheck
		default:
			logger.Debug("update: %#v", update)
		}

		return true
	})

	go func() {
		logger.Info("starting receive updates...")
		t.bot.Start()
	}()

	// Block until context done
	<-t.ctx.Done()

	return ErrServiceClosed
}

// Shutdown shuts down the Telegram service, it always retuan a nil error.
func (t *Telegram) Shutdown() error {
	if t.bot != nil {
		t.bot.Stop()
	}

	return nil
}

// nolint:gocyclo
func (t *Telegram) process(message *telegram.Message) (err error) {
	content := message.Text
	logger.Debug("content: %s", content)

	// If the message is forwarded and contains multiple entities,
	// the update will be split into multiple parts.
	// Don't process parts of the forwarded message without text.
	// if message.IsForwarded() && message.Caption == "" {
	if message.IsForwarded() && content == "" {
		return nil
	}
	urls := service.ExcludeURL(service.MatchURL(t.opts, content), "t.me")

	// Set command as playback if receive a playback command without URLs, and
	// required user reply a message with URLs.
	if message.IsReply() {
		if message.ReplyTo.Sender.Username == t.bot.Me.Username {
			content = "/playback" + content
		}
	}

	command := command(content)
	switch {
	case command == service.CommandHelp, command == "start":
		// nolint:errcheck
		t.reply(message, t.opts.TelegramHelptext())
	case command == service.CommandPlayback:
		return t.playback(message)
	case command == service.CommandMetrics:
		stats := metrics.Gather.Export("wayback")
		if t.opts.EnabledMetrics() && stats != "" {
			if _, err = t.reply(message, stats); err != nil {
				return err
			}
		}
		return nil
	case command == service.CommandPrivacy:
		// nolint:errcheck
		t.reply(message, fmt.Sprintf("To read our privacy policy, please visit %s.", t.opts.PrivacyURL()))
	case command != "":
		fallback := t.commandFallback()
		if fallback != "" {
			fallback = fmt.Sprintf("\n\nAvailable commands:\n%s", fallback)
		}
		// nolint:errcheck
		t.reply(message, fmt.Sprintf("/%s is an illegal command%s", command, fallback))
	case len(urls) == 0:
		logger.Warn("archives failure, URL no found.")
		metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusRequest)
		t.reply(message, "URL no found.") // nolint:errcheck
	default:
		metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusRequest)
		request, err := t.reply(message, "Queue...")
		if err != nil {
			return errors.Wrap(err, "reply message failed")
		}
		bucket := pooling.Bucket{
			Request: func(ctx context.Context) error {
				_, err := t.bot.Edit(request, "Archiving...")
				if err != nil && err != telegram.ErrSameMessageContent {
					return errors.Wrap(err, "telegram: send archiving message failed")
				}

				if err := t.wayback(ctx, request, urls); err != nil {
					// nolint:errcheck
					t.bot.Edit(request, service.MsgWaybackRetrying)
					return errors.Wrap(err, "archives failed")
				}
				metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusSuccess)
				return nil
			},
			Fallback: func(_ context.Context) error {
				t.bot.Delete(request)                           // nolint:errcheck
				t.bot.Reply(message, service.MsgWaybackTimeout) // nolint:errcheck
				metrics.IncrementWayback(metrics.ServiceTelegram, metrics.StatusFailure)
				return nil
			},
		}
		t.pool.Put(bucket)
	}
	return nil
}

func (t *Telegram) wayback(ctx context.Context, request *telegram.Message, urls []*url.URL) error {
	do := func(cols []wayback.Collect, rdx reduxer.Reduxer) error {
		opts := &telegram.SendOptions{DisableWebPagePreview: true}
		replyText := render.ForReply(&render.Telegram{Cols: cols, Data: rdx}).String()
		logger.Debug("reply text, %s", replyText)

		if _, err := t.bot.Edit(request, replyText, opts); err != nil {
			return errors.Wrap(err, "telegram: update message failed")
		}

		t.pub.Spread(ctx, rdx, cols, publish.FlagTelegram)

		var albums telegram.Album
		var head = render.Title(cols, rdx)

		for _, u := range urls {
			if b, ok := rdx.Load(reduxer.Src(u.String())); ok {
				albums = append(albums, service.UploadToTelegram(t.opts, b.Artifact(), head)...)
			}
		}
		if len(albums) == 0 {
			logger.Debug("no albums to send")
			return nil
		}

		// Send album attach files, and reply to wayback result message
		opts = &telegram.SendOptions{ReplyTo: request, DisableNotification: true}
		if _, err := t.bot.SendAlbum(request.Chat, albums, opts); err != nil {
			logger.Error("reply failed: %v", err)
		}
		return nil
	}

	return service.Wayback(ctx, t.opts, urls, do)
}

func (t *Telegram) playback(message *telegram.Message) error {
	metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusRequest)

	recipient, err := t.bot.ChatByID(message.Chat.ID)
	if err != nil {
		metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusFailure)
		logger.Error("playback failed: %v", err)
		return err
	}

	urls := service.MatchURL(t.opts, message.Text)
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
	cols, err := wayback.Playback(t.ctx, t.opts, urls...)
	if err != nil {
		return errors.Wrap(err, "telegram: playback failed")
	}
	logger.Debug("playback collections: %#v", cols)

	// Due to Telegram restricted callback data to 1-64 bytes, it requires to store
	// playback URLs to database.
	data := helper.String2Byte(strings.ReplaceAll(callbackPrefix()+message.Text, "/playback", ""))
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
					Data: strconv.FormatUint(pb.ID, 10),
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
	commands, err := t.bot.Commands()
	if err != nil {
		logger.Error("got my commands failed: %v", err)
	}

	var maps = make(map[string]bool, len(commands))
	for _, command := range commands {
		maps[command.Text] = true
	}

	for _, command := range t.defaultCommands() {
		if maps[command.Text] {
			continue
		}
		commands = append(commands, command)
	}

	return commands
}

func (t *Telegram) setCommands() error {
	commands := t.getCommands()
	logger.Debug("got commands: %v", commands)

	if err := t.bot.SetCommands(commands); err != nil {
		logger.Error("set commands failed: %v", err)
		return err
	}
	logger.Debug("set commands succeed")

	return nil
}

func (t *Telegram) defaultCommands() []telegram.Command {
	commands := []telegram.Command{
		{
			Text:        service.CommandHelp,
			Description: "Show help information",
		},
		{
			Text:        service.CommandPlayback,
			Description: "Playback archived url",
		},
	}
	if t.opts.PrivacyURL() != "" {
		commands = append(commands, telegram.Command{
			Text:        service.CommandPrivacy,
			Description: "Read our privacy policy",
		})
	}
	if t.opts.EnabledMetrics() {
		commands = append(commands, telegram.Command{
			Text:        service.CommandMetrics,
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
		return service.CommandHelp
	case strings.HasPrefix(message, "/playback"):
		return service.CommandPlayback
	case strings.HasPrefix(message, "/metrics"):
		return service.CommandMetrics
	case strings.HasPrefix(message, "/privacy"):
		return service.CommandPrivacy
	default:
		return matchCmd(message)
	}
}

func transform(m *telegram.Message) {
	entities := func(e telegram.Entities) (uri []string) {
		for _, entity := range e {
			if entity.URL != "" {
				uri = append(uri, entity.URL)
			}
		}
		return
	}

	// At lease one embed link is included in the message.
	if len(m.Entities) > 0 {
		uri := entities(m.Entities)
		m.Text = fmt.Sprintf("%s and URI in message entity: %s", m.Text, strings.Join(uri, space))
	}
	// The message body is an attachment with a caption.
	if m.Caption != "" {
		m.Text = fmt.Sprintf("%s and caption: %s", m.Text, m.Caption)
	}
	if len(m.CaptionEntities) > 0 {
		uri := entities(m.CaptionEntities)
		m.Text = fmt.Sprintf("%s and URI in caption entity: %s", m.Text, strings.Join(uri, space))
	}
}
