// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"os"
	"path"
	"strings"

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
	var txt = render.ForPublish(&render.Discord{Cols: cols}).String()
	if d.toChannel(ctx, &bnd, txt) {
		metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishDiscord, metrics.StatusFailure)
	return
}

// toChannel for publish to message to Discord channel,
// returns boolean as result.
func (d *discordBot) toChannel(ctx context.Context, bundle *reduxer.Bundle, text string) (ok bool) {
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

	// TODO: move to render
	var b strings.Builder
	if head := title(ctx, bundle); head != "" {
		b.WriteString(`**`)
		b.WriteString(head)
		b.WriteString(`**`)
		b.WriteString("\n\n")
	}
	if dgst := digest(ctx, bundle); dgst != "" {
		b.WriteString(dgst)
		b.WriteString("\n\n")
	}
	b.WriteString(text)

	msg := &discord.MessageSend{Content: b.String()}
	if bundle != nil {
		var fsize int64
		var files []*discord.File
		upper := config.Opts.MaxAttachSize("discord")
		for _, p := range bundle.Paths() {
			if p == "" {
				continue
			}
			if !helper.Exists(p) {
				logger.Warn("invalid file %s", p)
				continue
			}
			fsize += helper.FileSize(p)
			if fsize > upper {
				logger.Warn("total file size large than %s, skipped", humanize.Bytes(uint64(upper)))
				continue
			}
			logger.Debug("open file: %s", p)
			rd, err := os.Open(p)
			if err != nil {
				logger.Error("open file failed: %v", err)
				continue
			}
			files = append(files, &discord.File{Name: path.Base(p), Reader: rd})
		}
		if len(files) == 0 {
			logger.Warn("files empty")
			return ok
		}
		msg.Files = files
	}

	_, err := d.bot.ChannelMessageSendComplex(config.Opts.DiscordChannel(), msg)
	if err != nil {
		logger.Error("post message to channel failed, %v", err)
		return ok
	}

	return true
}
