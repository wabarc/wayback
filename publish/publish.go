// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	mstdn "github.com/mattn/go-mastodon"
	irc "github.com/thoj/go-ircevent"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/reduxer"
	"golang.org/x/sync/errgroup"
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

	maxTitleLen  = 256
	maxDigestLen = 500
)

var maxDelayTime = 10

// Publisher is the interface that wraps the basic Publish method.
//
// Publish publish message to serveral media platforms, e.g. Telegram channel, GitHub Issues, etc.
// The cols must either be a []wayback.Collect, args use for specific service.
type Publisher interface {
	Publish(ctx context.Context, cols []wayback.Collect, args ...string)
}

func process(p Publisher, ctx context.Context, cols []wayback.Collect, args ...string) {
	// Compose the collects into multiple parts by URI
	var parts = make(map[string][]wayback.Collect)
	for _, col := range cols {
		parts[col.Src] = append(parts[col.Src], col)
	}

	f := from(args...)
	g, ctx := errgroup.WithContext(ctx)
	for _, part := range parts {
		logger.Debug("[%s] produce part: %#v", f, part)

		part := part
		g.Go(func() error {
			// Nice for target server. It should be skipped on the testing mode.
			if !strings.HasSuffix(os.Args[0], ".test") {
				rand.Seed(time.Now().UnixNano())
				r := rand.Intn(maxDelayTime) //nolint:gosec,goimports
				w := time.Duration(r) * time.Second
				logger.Debug("[%s] produce sleep %d second", f, r)
				time.Sleep(w)
			}

			p.Publish(ctx, part, args...)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.Error("[%s] produce failed: %v", f, err)
		return
	}

	return
}

func from(args ...string) (f string) {
	if len(args) > 0 {
		f = args[0]
	}
	return f
}

// nolint:gocyclo
func To(ctx context.Context, cols []wayback.Collect, args ...string) {
	f := from(args...)
	channel := func(ctx context.Context, cols []wayback.Collect, args ...string) {
		if config.Opts.PublishToChannel() {
			logger.Debug("[%s] publishing to telegram channel...", f)
			var bot *telegram.Bot
			if rev, ok := ctx.Value(FlagTelegram).(*telegram.Bot); ok {
				bot = rev
			}
			if bot == nil {
				return
			}
			t := NewTelegram(bot)
			process(t, ctx, cols, args...)
		}
	}
	issue := func(ctx context.Context, cols []wayback.Collect, args ...string) {
		if config.Opts.PublishToIssues() {
			logger.Debug("[%s] publishing to GitHub issues...", f)
			gh := NewGitHub(nil)
			process(gh, ctx, cols, args...)
		}
	}
	mastodon := func(ctx context.Context, cols []wayback.Collect, args ...string) {
		if config.Opts.PublishToMastodon() {
			logger.Debug("[%s] publishing to Mastodon...", f)
			var client *mstdn.Client
			if rev, ok := ctx.Value(FlagMastodon).(*mstdn.Client); ok {
				client = rev
			}
			mstdn := NewMastodon(client)
			process(mstdn, ctx, cols, args...)
		}
	}
	matrix := func(ctx context.Context, cols []wayback.Collect, args ...string) {
		if config.Opts.PublishToMatrixRoom() {
			logger.Debug("[%s] publishing to Matrix room...", f)
			var client *matrix.Client
			if rev, ok := ctx.Value(FlagMatrix).(*matrix.Client); ok {
				client = rev
			}
			mat := NewMatrix(client)
			process(mat, ctx, cols, args...)
		}
	}
	twitter := func(ctx context.Context, cols []wayback.Collect, args ...string) {
		if config.Opts.PublishToTwitter() {
			logger.Debug("[%s] publishing to Twitter...", f)
			var client *twitter.Client
			if rev, ok := ctx.Value(FlagTwitter).(*twitter.Client); ok {
				client = rev
			}
			twitter := NewTwitter(client)
			process(twitter, ctx, cols, args...)
		}
	}
	irc := func(ctx context.Context, cols []wayback.Collect, args ...string) {
		if config.Opts.PublishToIRCChannel() {
			logger.Debug("[%s] publishing to IRC channel...", f)
			var conn *irc.Connection
			if rev, ok := ctx.Value(FlagIRC).(*irc.Connection); ok {
				conn = rev
			}
			irc := NewIRC(conn)
			process(irc, ctx, cols, args...)
		}
	}
	funcs := map[string]func(context.Context, []wayback.Collect, ...string){
		"channel":  channel,
		"issue":    issue,
		"mastodon": mastodon,
		"matrix":   matrix,
		"twitter":  twitter,
		"irc":      irc,
	}

	g, ctx := errgroup.WithContext(ctx)
	for k, fn := range funcs {
		logger.Debug(`[%s] processing func %s`, f, k)
		fn := fn
		g.Go(func() error {
			fn(ctx, cols, args...)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.Error("[%s] process failed: %v", f, err)
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

func bundle(ctx context.Context, cols []wayback.Collect) (b reduxer.Bundle) {
	if len(cols) == 0 {
		return b
	}

	var uri = cols[0].Src
	if bundles, ok := ctx.Value(PubBundle).(reduxer.Bundles); ok {
		b = bundles[uri]
	}

	return b
}

func title(_ context.Context, bundle *reduxer.Bundle) string {
	if bundle == nil {
		return ""
	}
	logger.Debug("extract title from reduxer bundle title: %s", bundle.Title)

	t := []rune(bundle.Title)
	l := len(t)
	if l > maxTitleLen {
		t = t[:maxTitleLen]
	}

	return strings.TrimSpace(string(t))
}

func digest(_ context.Context, bundle *reduxer.Bundle) string {
	if bundle == nil {
		return ""
	}
	logger.Debug("generate digest from article content: %s", bundle.Article.TextContent)

	txt := []rune(bundle.Article.TextContent)
	l := len(txt)
	switch {
	case l == 0:
		return ""
	case l > maxDigestLen:
		txt = txt[:maxDigestLen]
		return string(txt) + ` ...`
	default:
		return string(txt)
	}
}
