// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package discord // import "github.com/wabarc/wayback/publish/discord"

import (
	"context"
	"net/http"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/template/render"

	discord "github.com/bwmarrin/discordgo"
)

// Interface guard
var _ publish.Publisher = (*Discord)(nil)

type Discord struct {
	bot  *discord.Session
	opts *config.Options
}

// New returns Discord bot client
func New(client *http.Client, opts *config.Options) *Discord {
	if !opts.PublishToDiscordChannel() {
		logger.Debug("Missing required environment variable, abort.")
		return nil
	}

	bot, err := discord.New("Bot " + opts.DiscordBotToken())
	if err != nil {
		logger.Error("create discord bot instance failed: %v", err)
		return nil
	}
	if client != nil {
		bot.Client = client
	}

	return &Discord{bot: bot, opts: opts}
}

// Publish publish text to the Discord channel of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (d *Discord) Publish(ctx context.Context, rdx reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusFailure)
		return errors.New("publish to discord: collects empty")
	}

	var body = render.ForPublish(&render.Discord{Cols: cols, Data: rdx}).String()
	if d.toChannel(rdx, body) {
		metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusSuccess)
		return nil
	}
	metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusFailure)
	return errors.New("publish to discord failed")
}

// toChannel for publish to message to Discord channel,
// returns boolean as result.
func (d *Discord) toChannel(rdx reduxer.Reduxer, body string) (ok bool) {
	if body == "" {
		logger.Warn("post to message to channel failed, body empty")
		return ok
	}
	if d.bot == nil {
		var err error
		d.bot, err = discord.New("Bot " + d.opts.DiscordBotToken())
		if err != nil {
			logger.Error("create discord bot instance failed: %v", err)
			return ok
		}
	}

	msg, err := d.bot.ChannelMessageSendComplex(d.opts.DiscordChannel(), &discord.MessageSend{Content: body})
	if err != nil {
		logger.Error("post message to channel failed, %v", err)
		return ok
	}

	// Send files as reference
	files, closeFunc := service.UploadToDiscord(d.opts, rdx)
	defer closeFunc()
	if len(files) == 0 {
		logger.Debug("without files, complete.")
		return true
	}

	ms := &discord.MessageSend{Files: files, Reference: msg.Reference()}
	if _, err := d.bot.ChannelMessageSendComplex(d.opts.DiscordChannel(), ms); err != nil {
		logger.Error("upload files failed, %v", err)
	}

	return true
}

// Shutdown shuts down the Discord publish service.
func (d *Discord) Shutdown() error {
	return d.bot.Close()
}
