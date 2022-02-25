// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/template/render"

	discord "github.com/bwmarrin/discordgo"
)

var _ Publisher = (*discordBot)(nil)

type discordBot struct {
	bot *discord.Session
}

// NewDiscord returns Discord bot client
func NewDiscord(bot *discord.Session) *discordBot {
	if !config.Opts.PublishToDiscordChannel() {
		logger.Error("Missing required environment variable, abort.")
		return new(discordBot)
	}

	if bot == nil {
		var err error
		bot, err = discord.New("Bot " + config.Opts.DiscordBotToken())
		if err != nil {
			logger.Error("create discord bot instance failed: %v", err)
		}
	}

	return &discordBot{bot: bot}
}

// Publish publish text to the Discord channel of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (d *discordBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusRequest)

	if len(cols) == 0 {
		return errors.New("publish to discord: collects empty")
	}

	rdx, art, err := extract(ctx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	var body = render.ForPublish(&render.Discord{Cols: cols, Data: rdx}).String()
	if d.toChannel(art, body) {
		metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusSuccess)
		return nil
	}
	metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusFailure)
	return errors.New("publish to discord failed")
}

// toChannel for publish to message to Discord channel,
// returns boolean as result.
func (d *discordBot) toChannel(art reduxer.Artifact, body string) (ok bool) {
	if body == "" {
		logger.Warn("post to message to channel failed, body empty")
		return ok
	}
	if d.bot == nil {
		var err error
		d.bot, err = discord.New("Bot " + config.Opts.DiscordBotToken())
		if err != nil {
			logger.Error("create discord bot instance failed: %v", err)
			return ok
		}
	}

	msg, err := d.bot.ChannelMessageSendComplex(config.Opts.DiscordChannel(), &discord.MessageSend{Content: body})
	if err != nil {
		logger.Error("post message to channel failed, %v", err)
		return ok
	}

	// Send files as reference
	files := service.UploadToDiscord(art)
	if len(files) == 0 {
		logger.Debug("without files, complete.")
		return true
	}
	ms := &discord.MessageSend{Files: files, Reference: msg.Reference()}
	if _, err := d.bot.ChannelMessageSendComplex(config.Opts.DiscordChannel(), ms); err != nil {
		logger.Error("upload files failed, %v", err)
	}

	return true
}
