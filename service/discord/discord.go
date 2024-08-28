// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package discord // import "github.com/wabarc/wayback/service/discord"

import (
	"context"
	"encoding/base64"
	"net/url"
	"strconv"
	"strings"

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

	discord "github.com/bwmarrin/discordgo"
)

// Interface guard
var _ service.Servicer = (*Discord)(nil)

// ErrServiceClosed is returned by the Service's Serve method after a call to Shutdown.
var ErrServiceClosed = errors.New("discord: Service closed")

// Discord represents a discord service in the application.
type Discord struct {
	ctx context.Context

	bot   *discord.Session
	store *storage.Storage
	opts  *config.Options
	pool  *pooling.Pool
	pub   *publish.Publish
}

// New returns a Discord struct.
func New(ctx context.Context, opts service.Options) (*Discord, error) {
	if !opts.Config.DiscordEnabled() {
		return nil, errors.New("missing required environment variable, skipped")
	}
	bot, err := discord.New("Bot " + opts.Config.DiscordBotToken())
	if err != nil {
		return nil, errors.Wrap(err, "create discord bot instance failed")
	}
	// Debug mode for bwmarrin/discordgo will print the bot token, should not apply it on production
	// if opts.Config.HasDebugMode() {
	//     bot.LogLevel = discord.LogDebug
	// }

	if ctx == nil {
		ctx = context.Background()
	}

	return &Discord{
		ctx:   ctx,
		bot:   bot,
		store: opts.Storage,
		opts:  opts.Config,
		pool:  opts.Pool,
		pub:   opts.Publish,
	}, nil
}

// Serve loop request message from the Discord api server.
// Serve always returns an error.
func (d *Discord) Serve() (err error) {
	if d.bot == nil {
		return errors.New("Initialize discord failed, error: %v", err)
	}
	d.bot.AddHandler(func(s *discord.Session, _ *discord.Ready) {
		logger.Info("authorized on account %s", color.Blue.Sprint(s.State.User.Username))
	})

	if channel, err := d.bot.UserChannelCreate(d.opts.DiscordChannel()); err == nil {
		logger.Info("channel name: %s, channel id: %s", color.Blue.Sprint(channel.Name), color.Blue.Sprint(channel.ID))
	}

	commandHandlers := d.commandHandlers()
	buttonHandlers := d.buttonHandlers()
	d.bot.AddHandler(func(s *discord.Session, i *discord.InteractionCreate) {
		switch i.Type {
		case discord.InteractionMessageComponent:
			// Type for button press will be always InteractionButton (3)
			// For playback
			if h, ok := buttonHandlers["playback"]; ok {
				h(s, i)
			}
		case discord.InteractionApplicationCommand:
			// Handle command
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		default:
			logger.Warn("skip %v", i.Type)
		}
	})

	// Handle message
	d.bot.AddHandler(func(s *discord.Session, m *discord.MessageCreate) {
		logger.Debug("received message create event: %#v", m.Message)
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}
		// Reply message and mention bot on the channel
		ref := m.Message.MessageReference
		if ref != nil {
			if msg, err := d.bot.ChannelMessage(ref.ChannelID, ref.MessageID); err != nil {
				logger.Debug("received message reference event error: %v", err)
			} else {
				logger.Debug("received message reference event: %#v", msg)
				m.Message.Content += msg.Content
			}
		}
		// nolint:errcheck
		d.process(m)
	})

	// Handle guild create event
	d.bot.AddHandler(func(s *discord.Session, g *discord.GuildCreate) {
		logger.Debug("guild: %#v", g.Guild)
		// d.setCommands(g.Guild.ID)
	})

	d.bot.AddHandler(func(s *discord.Session, _ *discord.Ready) {
		logger.Debug("set global commands")
		// Set global bot commands
		// nolint:errcheck
		d.setCommands("")
	})

	go func() {
		logger.Info("starting receive updates...")
		if err := d.bot.Open(); err != nil {
			logger.Error(`open connection failed: %v`, err)
		}
	}()

	// Block until context done
	<-d.ctx.Done()

	return ErrServiceClosed
}

// Shutdown shuts down the Discord service.
func (d *Discord) Shutdown() error {
	if d.bot != nil {
		return d.bot.Close()
	}

	return nil
}

func (d *Discord) commandHandlers() map[string]func(*discord.Session, *discord.InteractionCreate) {
	return map[string]func(s *discord.Session, i *discord.InteractionCreate){
		"help": func(s *discord.Session, i *discord.InteractionCreate) {
			// nolint:errcheck
			s.InteractionRespond(i.Interaction, &discord.InteractionResponse{
				Type: discord.InteractionResponseChannelMessageWithSource,
				Data: &discord.InteractionResponseData{
					Content: d.opts.DiscordHelptext(),
				},
			})
		},
		"playback": func(s *discord.Session, i *discord.InteractionCreate) {
			d.playback(s, i) // nolint:errcheck
		},
		"metrics": func(s *discord.Session, i *discord.InteractionCreate) {
			stats := metrics.Gather.Export("wayback")
			if !d.opts.EnabledMetrics() || stats == "" {
				return
			}
			// nolint:errcheck
			s.InteractionRespond(i.Interaction, &discord.InteractionResponse{
				Type: discord.InteractionResponseChannelMessageWithSource,
				Data: &discord.InteractionResponseData{
					Content: stats,
				},
			})
		},
	}
}

func (d *Discord) buttonHandlers() map[string]func(*discord.Session, *discord.InteractionCreate) {
	return map[string]func(s *discord.Session, i *discord.InteractionCreate){
		"playback": func(s *discord.Session, i *discord.InteractionCreate) {
			id, err := strconv.Atoi(i.MessageComponentData().CustomID)
			if err != nil {
				logger.Warn("invalid playback id: %s", i.MessageComponentData().CustomID)
				metrics.IncrementWayback(metrics.ServiceDiscord, metrics.StatusFailure)
				return
			}

			// Query playback callback data from database
			pb, err := d.store.Playback(uint64(id))
			if err != nil {
				logger.Error("query playback data failed: %v", err)
				metrics.IncrementWayback(metrics.ServiceDiscord, metrics.StatusFailure)
				return
			}

			data, err := base64.StdEncoding.DecodeString(pb.Source)
			if err != nil {
				logger.Error("decoding callback data failed: %v", err)
				metrics.IncrementWayback(metrics.ServiceDiscord, metrics.StatusFailure)
				return
			}

			// Send an interaction respond to markup interact status
			// nolint:errcheck
			s.InteractionRespond(i.Interaction, &discord.InteractionResponse{
				Type: discord.InteractionResponseChannelMessageWithSource,
				Data: &discord.InteractionResponseData{
					Content: "Processing...",
				},
			})

			// nolint:errcheck
			s.ChannelTyping(i.Message.ChannelID)

			i.Message.Content = helper.Byte2String(data)
			d.process(&discord.MessageCreate{Message: i.Message})       // nolint:errcheck
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction) // nolint:errcheck
		},
	}
}

// nolint:gocyclo
func (d *Discord) process(m *discord.MessageCreate) (err error) {
	content := m.Content
	logger.Debug("content: %s", content)

	urls := service.MatchURL(d.opts, content)

	switch {
	case m.GuildID != "" && !d.isMention(content):
		// don't process message from channel and without mention
		logger.Debug("message from channel and without mention, skipped")
	case len(urls) == 0:
		logger.Warn("archives failure, URL no found.")
		metrics.IncrementWayback(metrics.ServiceDiscord, metrics.StatusRequest)
		d.reply(m, "URL no found.") // nolint:errcheck
	default:
		metrics.IncrementWayback(metrics.ServiceDiscord, metrics.StatusRequest)
		if m, err = d.reply(m, "Queue..."); err != nil {
			logger.Error("reply queue failed: %v", err)
			return
		}
		bucket := pooling.Bucket{
			Request: func(ctx context.Context) error {
				logger.Debug("content: %v", urls)
				if err := d.wayback(ctx, m, urls); err != nil {
					logger.Error("archives failed: %v", err)
					// nolint:errcheck
					d.reply(m, service.MsgWaybackRetrying)
					return err
				}
				metrics.IncrementWayback(metrics.ServiceDiscord, metrics.StatusSuccess)
				return nil
			},
			Fallback: func(_ context.Context) error {
				// nolint:errcheck
				d.reply(m, service.MsgWaybackTimeout)
				metrics.IncrementWayback(metrics.ServiceDiscord, metrics.StatusFailure)
				return nil
			},
		}
		d.pool.Put(bucket)
	}
	return nil
}

func (d *Discord) wayback(ctx context.Context, m *discord.MessageCreate, urls []*url.URL) error {
	stage, err := d.edit(m, "Archiving...")
	if err != nil {
		logger.Error("send archiving message failed: %v", err)
		return err
	}
	logger.Debug("send archiving message result: %#v", stage)

	do := func(cols []wayback.Collect, rdx reduxer.Reduxer) error {
		replyText := render.ForReply(&render.Discord{Cols: cols}).String()
		logger.Debug("reply text, %s", replyText)

		if _, err := d.edit(stage, replyText); err != nil {
			return errors.Wrap(err, "discord: update message failed")
		}

		// Avoid republishing
		if m.ChannelID != d.opts.DiscordChannel() {
			d.pub.Spread(ctx, rdx, cols, publish.FlagDiscord)
		}

		msg := &discord.MessageSend{Content: replyText, Reference: stage.Message.Reference()}
		var files []*discord.File
		for _, u := range urls {
			if bundle, ok := rdx.Load(reduxer.Src(u.String())); ok {
				files = append(files, service.UploadToDiscord(d.opts, bundle.Artifact())...)
			}
		}
		if len(files) == 0 {
			logger.Warn("files empty")
			return nil
		}
		msg.Files = files

		if _, err := d.bot.ChannelMessageSendComplex(m.ChannelID, msg); err != nil {
			logger.Error("post message to channel failed, %v", err)
			return err
		}
		return nil
	}

	return service.Wayback(ctx, d.opts, urls, do)
}

func (d *Discord) playback(s *discord.Session, i *discord.InteractionCreate) error {
	metrics.IncrementPlayback(metrics.ServiceDiscord, metrics.StatusRequest)

	text := i.ApplicationCommandData().Options[0].StringValue()
	urls := service.MatchURL(d.opts, text)
	if len(urls) == 0 {
		return d.bot.InteractionRespond(i.Interaction, &discord.InteractionResponse{
			Type: discord.InteractionResponseChannelMessageWithSource,
			Data: &discord.InteractionResponseData{
				Content: "Please send me URLs to playback...",
			},
		})
	}

	// nolint:errcheck
	s.InteractionRespond(i.Interaction, &discord.InteractionResponse{
		Type: discord.InteractionResponseChannelMessageWithSource,
		Data: &discord.InteractionResponseData{
			Content: "Processing...",
		},
	})

	cols, err := wayback.Playback(d.ctx, d.opts, urls...)
	if err != nil {
		return errors.Wrap(err, "discord: playback failed")
	}
	logger.Debug("playback collections: %#v", cols)

	// Due to Discord restricted custom_id up to 100 characters, it requires to store
	// playback URLs to database.
	pb := &entity.Playback{Source: base64.StdEncoding.EncodeToString(helper.String2Byte(text))}
	if err = d.store.CreatePlayback(pb); err != nil {
		logger.Error("store collections failed: %v", err)
		return err
	}

	replyText := render.ForReply(&render.Discord{Cols: cols}).String()
	err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discord.WebhookEdit{
		Content: replyText,
		Components: []discord.MessageComponent{
			discord.ActionsRow{
				Components: []discord.MessageComponent{
					discord.Button{
						Label:    "wayback",
						Style:    discord.SuccessButton,
						Disabled: false,
						CustomID: strconv.FormatUint(pb.ID, 10),
					},
				},
			},
		},
	})
	if err != nil {
		metrics.IncrementPlayback(metrics.ServiceDiscord, metrics.StatusFailure)
		logger.Error("send playback results failed: %v", err)
		return err
	}
	metrics.IncrementPlayback(metrics.ServiceDiscord, metrics.StatusSuccess)
	return nil
}

func (d *Discord) reply(m *discord.MessageCreate, text string) (*discord.MessageCreate, error) {
	if text == "" {
		logger.Warn("text empty, skipped")
		return nil, errors.New("text empty")
	}

	var err error
	m.Message, err = d.bot.ChannelMessageSendReply(m.Message.ChannelID, text, m.Message.Reference())
	if err != nil {
		logger.Error("reply failed: %v", err)
		return m, err
	}
	return m, nil
}

func (d *Discord) edit(m *discord.MessageCreate, text string) (*discord.MessageCreate, error) {
	if text == "" {
		logger.Warn("text empty, skipped")
		return nil, errors.New("text empty")
	}

	var err error
	m.Message, err = d.bot.ChannelMessageEdit(m.ChannelID, m.Message.ID, text)
	if err != nil {
		logger.Error("edit failed: %v", err)
		return m, err
	}
	return m, nil
}

func (d *Discord) setCommands(guild string) (err error) {
	if _, err = d.bot.ApplicationCommandBulkOverwrite(d.bot.State.User.ID, guild, d.requires()); err != nil {
		logger.Error("overwrite commands failed: %v", err)
		return err
	}
	logger.Info("set commands succeed")

	return nil
}

func (d *Discord) requires() (commands []*discord.ApplicationCommand) {
	if d.opts.DiscordHelptext() != "" {
		commands = append(commands, &discord.ApplicationCommand{
			Name:        "help",
			Description: "Show help information",
		})
	}
	if d.opts.EnabledMetrics() {
		commands = append(commands, &discord.ApplicationCommand{
			Name:        "metrics",
			Description: "Show service metrics",
		})
	}
	commands = append(commands, &discord.ApplicationCommand{
		Name:        "playback",
		Description: "Playback archived url",
		Options: []*discord.ApplicationCommandOption{
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        "urls",
				Description: "Send me URLs to playback...",
				Required:    true,
			},
		},
	})

	return commands
}

func (d *Discord) isMention(content string) bool {
	prefix := "<@!" + d.bot.State.User.ID + ">"
	return strings.HasPrefix(content, prefix)
}
