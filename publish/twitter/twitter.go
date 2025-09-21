// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package twitter // import "github.com/wabarc/wayback/publish/twitter"

import (
	"context"
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"
)

// Interface guard
var _ publish.Publisher = (*Twitter)(nil)

type Twitter struct {
	ctx context.Context

	bot  *twitter.Client
	opts *config.Options
}

// New returns a twitter client.
func New(ctx context.Context, client *http.Client, opts *config.Options) *Twitter {
	if !opts.PublishToTwitter() {
		logger.Debug("Missing required environment variable")
		return nil
	}

	if client == nil {
		oauth := oauth1.NewConfig(opts.TwitterConsumerKey(), opts.TwitterConsumerSecret())
		token := oauth1.NewToken(opts.TwitterAccessToken(), opts.TwitterAccessSecret())
		client = oauth.Client(oauth1.NoContext, token)
	}
	bot := twitter.NewClient(client)

	return &Twitter{ctx: ctx, bot: bot, opts: opts}
}

// Publish publish tweet to Twitter of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (t *Twitter) Publish(ctx context.Context, rdx reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishTwitter, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishTwitter, metrics.StatusFailure)
		return errors.New("publish to twitter: collects empty")
	}

	_, err := publish.Artifact(ctx, rdx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	var body = render.ForPublish(&render.Twitter{Cols: cols, Data: rdx}).String()
	if t.ToTwitter(ctx, body) {
		metrics.IncrementPublish(metrics.PublishTwitter, metrics.StatusSuccess)
		return nil
	}
	metrics.IncrementPublish(metrics.PublishTwitter, metrics.StatusFailure)
	return errors.New("publish to twitter failed")
}

func (t *Twitter) ToTwitter(ctx context.Context, body string) bool {
	if !t.opts.PublishToTwitter() || t.bot == nil {
		logger.Warn("Do not publish to Twitter.")
		return false
	}
	if body == "" {
		logger.Warn("twitter validation failed: body can't be blank")
		return false
	}

	tweet, resp, err := t.bot.Statuses.Update(body, nil)
	if err != nil {
		logger.Error("create tweet failed: %v", err)
		return false
	}
	defer resp.Body.Close()
	logger.Debug("created tweet: %v", tweet)

	return true
}

// Shutdown shuts down the Twitter publish service, it always return a nil error.
func (t *Twitter) Shutdown() error {
	return nil
}
