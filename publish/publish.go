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
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	matrix "maunium.net/go/mautrix"
)

func To(ctx context.Context, col []*wayback.Collect, args ...string) {
	var from string
	if len(args) > 0 {
		from = args[0]
	}

	switch {
	case config.Opts.PublishToChannel():
		logger.Debug("[%s] publishing to channel...", from)

		var bot *telegram.BotAPI
		if rev, ok := ctx.Value("telegram").(*telegram.BotAPI); ok {
			bot = rev
		}
		tel := NewTelegram(bot)
		tel.ToChannel(ctx, tel.Render(col))
	case config.Opts.PublishToIssues():
		logger.Debug("[%s] publishing to GitHub issues...", from)
		gh := NewGitHub(nil)
		gh.ToIssues(ctx, gh.Render(col))
	case config.Opts.PublishToMastodon():
		var id string
		if len(args) > 1 {
			id = args[1]
		}
		logger.Debug("[%s] publishing to Mastodon...", from)

		var client *mstdn.Client
		if rev, ok := ctx.Value("mastodon").(*mstdn.Client); ok {
			client = rev
		}
		mstdn := NewMastodon(client)
		mstdn.ToMastodon(ctx, mstdn.Render(col), id)
	case config.Opts.PublishToTwitter():
		logger.Debug("[%s] publishing to Twitter...", from)

		var client *twitter.Client
		if rev, ok := ctx.Value("twitter").(*twitter.Client); ok {
			client = rev
		}
		twitter := NewTwitter(client)
		twitter.ToTwitter(ctx, twitter.Render(col))
	case config.Opts.PublishToIRCChannel():
		logger.Debug("[%s] publishing to IRC channel...", from)

		var conn *irc.Connection
		if rev, ok := ctx.Value("irc").(*irc.Connection); ok {
			conn = rev
		}
		irc := NewIRC(conn)
		irc.ToChannel(ctx, irc.Render(col))
	case config.Opts.PublishToMatrixRoom():
		logger.Debug("[%s] publishing to Matrix room...", from)

		var client *matrix.Client
		if rev, ok := ctx.Value("matrix").(*matrix.Client); ok {
			client = rev
		}
		mat := NewMatrix(client)
		mat.ToRoom(ctx, mat.Render(col))
	}
}
