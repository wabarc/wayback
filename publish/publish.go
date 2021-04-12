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

		var bot *telegram.BotAPI
		if rev, ok := ctx.Value("telegram").(*telegram.BotAPI); ok {
			bot = rev
		}
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

		var client *mstdn.Client
		if rev, ok := ctx.Value("mastodon").(*mstdn.Client); ok {
			client = rev
		}
		mstdn := NewMastodon(client, opts)
		mstdn.ToMastodon(ctx, opts, mstdn.Render(col), id)
	}
	if opts.PublishToTwitter() {
		logger.Debug("[%s] publishing to Twitter...", from)

		var client *twitter.Client
		if rev, ok := ctx.Value("twitter").(*twitter.Client); ok {
			client = rev
		}
		twitter := NewTwitter(client, opts)
		twitter.ToTwitter(ctx, opts, twitter.Render(col))
	}
	if opts.PublishToIRCChannel() {
		logger.Debug("[%s] publishing to IRC channel...", from)

		var conn *irc.Connection
		if rev, ok := ctx.Value("irc").(*irc.Connection); ok {
			conn = rev
		}
		irc := NewIRC(conn, opts)
		irc.ToChannel(ctx, opts, irc.Render(col))
	}
	if opts.PublishToMatrixRoom() {
		logger.Debug("[%s] publishing to Matrix room...", from)

		var client *matrix.Client
		if rev, ok := ctx.Value("matrix").(*matrix.Client); ok {
			client = rev
		}
		matrix := NewMatrix(client, opts)
		matrix.ToRoom(ctx, opts, matrix.Render(col))
	}
}
