// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"

	"github.com/dghubble/go-twitter/twitter"
	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	mstdn "github.com/mattn/go-mastodon"
	irc "github.com/thoj/go-ircevent"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/logger"
	matrix "maunium.net/go/mautrix"
)

func To(ctx context.Context, opts *config.Options, col []*wayback.Collect, args ...string) {
	var from string
	if len(args) > 0 {
		from = args[0]
	}

	if opts.PublishToChannel() {
		logger.Debug("[%s] publishing to channel...", from)
		bot := ctx.Value("telegram").(*telegram.BotAPI)
		ToChannel(opts, bot, Render(col))
	}
	if opts.PublishToIssues() {
		logger.Debug("[%s] publishing to GitHub issues...", from)
		ToIssues(ctx, opts, NewGitHub().Render(col))
	}
	if opts.PublishToMastodon() {
		var id string
		if len(args) > 1 {
			id = args[1]
		}
		logger.Debug("[%s] publishing to Mastodon...", from)
		client := ctx.Value("mastodon").(*mstdn.Client)
		mstdn := NewMastodon(client, opts)
		mstdn.ToMastodon(ctx, opts, mstdn.Render(col), id)
	}
	if opts.PublishToTwitter() {
		logger.Debug("[%s] publishing to Twitter...", from)
		client := ctx.Value("twitter").(*twitter.Client)
		twitter := NewTwitter(client, opts)
		twitter.ToTwitter(ctx, opts, twitter.Render(col))
	}
	if opts.PublishToIRCChannel() {
		logger.Debug("[%s] publishing to IRC channel...", from)
		conn := ctx.Value("irc").(*irc.Connection)
		irc := NewIRC(conn, opts)
		irc.ToChannel(ctx, opts, irc.Render(col))
	}
	if opts.PublishToMatrixRoom() {
		logger.Debug("[%s] publishing to Matrix room...", from)
		client := ctx.Value("matrix").(*matrix.Client)
		matrix := NewMatrix(client, opts)
		matrix.ToRoom(ctx, opts, matrix.Render(col))
	}
}
