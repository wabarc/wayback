// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"os"
	"path"

	"github.com/dustin/go-humanize"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"

	discord "github.com/bwmarrin/discordgo"
)

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

func (d *discordBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Warn("collects empty")
		return
	}

	var bnd = bundle(ctx, cols)
	var txt = render.ForPublish(&render.Discord{Cols: cols, Data: bnd}).String()
	if d.toChannel(ctx, bnd, txt) {
		metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusFailure)
	return
}

// toChannel for publish to message to Discord channel,
// returns boolean as result.
func (d *discordBot) toChannel(_ context.Context, bundle *reduxer.Bundle, text string) (ok bool) {
	if text == "" {
		logger.Warn("post to message to channel failed, text empty")
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

	msg, err := d.bot.ChannelMessageSendComplex(config.Opts.DiscordChannel(), &discord.MessageSend{Content: text})
	if err != nil {
		logger.Error("post message to channel failed, %v", err)
		return ok
	}

	// Send files as reference
	files := UploadToDiscord(bundle)
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

func UploadToDiscord(bundle *reduxer.Bundle) (files []*discord.File) {
	if bundle != nil {
		var fsize int64
		upper := config.Opts.MaxAttachSize("discord")
		for _, asset := range bundle.Asset() {
			if asset.Local == "" {
				continue
			}
			if !helper.Exists(asset.Local) {
				logger.Warn("invalid file %s", asset.Local)
				continue
			}
			fsize += helper.FileSize(asset.Local)
			if fsize > upper {
				logger.Warn("total file size large than %s, skipped", humanize.Bytes(uint64(upper)))
				continue
			}
			logger.Debug("open file: %s", asset.Local)
			rd, err := os.Open(asset.Local)
			if err != nil {
				logger.Error("open file failed: %v", err)
				continue
			}
			files = append(files, &discord.File{Name: path.Base(asset.Local), Reader: rd})
		}
	}
	return
}
