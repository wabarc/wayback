// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"net/url"
	"strings"
	"text/template"

	"github.com/dghubble/go-twitter/twitter"
	mstdn "github.com/mattn/go-mastodon"
	irc "github.com/thoj/go-ircevent"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/metrics"
	telegram "gopkg.in/tucnak/telebot.v2"
	matrix "maunium.net/go/mautrix"
)

const (
	FlagWeb      = "web"
	FlagTelegram = "telegram"
	FlagTwitter  = "twitter"
	FlagMastodon = "mastodon"
	FlagMatrix   = "matrix"
	FlagIRC      = "irc"

	PubBundle = "reduxer-bundle"
)

// nolint:gocyclo
func To(ctx context.Context, col []*wayback.Collect, args ...string) {
	var from string
	if len(args) > 0 {
		from = args[0]
	}

	if config.Opts.PublishToChannel() {
		logger.Debug("[%s] publishing to telegram channel...", from)
		metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusRequest)

		var bot *telegram.Bot
		if rev, ok := ctx.Value(FlagTelegram).(*telegram.Bot); ok {
			bot = rev
		}

		tel := NewTelegram(bot)
		if tel.ToChannel(ctx, tel.Render(col)) {
			metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusSuccess)
		} else {
			metrics.IncrementPublish(metrics.PublishChannel, metrics.StatusFailure)
		}
	}
	if config.Opts.PublishToIssues() {
		logger.Debug("[%s] publishing to GitHub issues...", from)
		metrics.IncrementPublish(metrics.PublishGithub, metrics.StatusRequest)

		gh := NewGitHub(nil)
		if gh.ToIssues(ctx, gh.Render(col)) {
			metrics.IncrementPublish(metrics.PublishGithub, metrics.StatusSuccess)
		} else {
			metrics.IncrementPublish(metrics.PublishGithub, metrics.StatusFailure)
		}
	}
	if config.Opts.PublishToMastodon() {
		var id string
		if len(args) > 1 {
			id = args[1]
		}
		logger.Debug("[%s] publishing to Mastodon...", from)
		metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusRequest)

		var client *mstdn.Client
		if rev, ok := ctx.Value(FlagMastodon).(*mstdn.Client); ok {
			client = rev
		}
		mstdn := NewMastodon(client)
		if mstdn.ToMastodon(ctx, mstdn.Render(col), id) {
			metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusSuccess)
		} else {
			metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusFailure)
		}
	}
	if config.Opts.PublishToTwitter() {
		logger.Debug("[%s] publishing to Twitter...", from)
		metrics.IncrementPublish(metrics.PublishTwitter, metrics.StatusRequest)

		var client *twitter.Client
		if rev, ok := ctx.Value(FlagTwitter).(*twitter.Client); ok {
			client = rev
		}
		twitter := NewTwitter(client)
		if twitter.ToTwitter(ctx, twitter.Render(col)) {
			metrics.IncrementPublish(metrics.PublishTwitter, metrics.StatusSuccess)
		} else {
			metrics.IncrementPublish(metrics.PublishTwitter, metrics.StatusFailure)
		}
	}
	if config.Opts.PublishToIRCChannel() {
		logger.Debug("[%s] publishing to IRC channel...", from)
		metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusRequest)

		var conn *irc.Connection
		if rev, ok := ctx.Value(FlagIRC).(*irc.Connection); ok {
			conn = rev
		}
		irc := NewIRC(conn)
		if irc.ToChannel(ctx, irc.Render(col)) {
			metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusSuccess)
		} else {
			metrics.IncrementPublish(metrics.PublishIRC, metrics.StatusFailure)
		}
	}
	if config.Opts.PublishToMatrixRoom() {
		logger.Debug("[%s] publishing to Matrix room...", from)
		metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusRequest)

		var client *matrix.Client
		if rev, ok := ctx.Value(FlagMatrix).(*matrix.Client); ok {
			client = rev
		}
		mat := NewMatrix(client)
		if mat.ToRoom(ctx, mat.Render(col)) {
			metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusSuccess)
		} else {
			metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusFailure)
		}
	}
}

func funcMap() template.FuncMap {
	cache := "https://webcache.googleusercontent.com/search?q=cache:"
	return template.FuncMap{
		"unescape": func(link string) string {
			unescaped, err := url.QueryUnescape(link)
			if err != nil {
				return link
			}
			return unescaped
		},
		"isURL": helper.IsURL,
		"revert": func(link string) string {
			return strings.Replace(link, cache, "", 1)
		},
		"not": func(text, s string) bool {
			return !strings.Contains(text, s)
		},
	}
}
