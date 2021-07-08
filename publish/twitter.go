// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"
)

type twitterBot struct {
	client *twitter.Client
}

func NewTwitter(client *twitter.Client) *twitterBot {
	if !config.Opts.PublishToTwitter() {
		logger.Error("Missing required environment variable")
		return new(twitterBot)
	}

	if client == nil {
		oauth := oauth1.NewConfig(config.Opts.TwitterConsumerKey(), config.Opts.TwitterConsumerSecret())
		token := oauth1.NewToken(config.Opts.TwitterAccessToken(), config.Opts.TwitterAccessSecret())
		httpClient := oauth.Client(oauth1.NoContext, token)
		client = twitter.NewClient(httpClient)
	}

	return &twitterBot{client: client}
}

func (t *twitterBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	metrics.IncrementPublish(metrics.PublishTwitter, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Warn("[publish] collects empty")
		return
	}

	var bnd = bundle(ctx, cols)
	var txt = render.ForPublish(&render.Twitter{Cols: cols}).String()
	if t.ToTwitter(ctx, &bnd, txt) {
		metrics.IncrementPublish(metrics.PublishTwitter, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishTwitter, metrics.StatusFailure)
	return
}

func (t *twitterBot) ToTwitter(ctx context.Context, bundle *reduxer.Bundle, text string) bool {
	if !config.Opts.PublishToTwitter() || t.client == nil {
		logger.Warn("[publish] Do not publish to Twitter.")
		return false
	}
	if text == "" {
		logger.Warn("[publish] twitter validation failed: Text can't be blank")
		return false
	}

	// TODO: character limit
	var b strings.Builder
	if head := title(ctx, bundle); head != "" {
		b.WriteString(`‹ `)
		b.WriteString(head)
		b.WriteString(" ›\n\n")
	}
	b.WriteString(text)
	tweet, resp, err := t.client.Statuses.Update(b.String(), nil)
	if err != nil {
		logger.Error("[publish] create tweet failed: %v", err)
		return false
	}
	logger.Debug("[publish] created tweet: %v, resp: %v, err: %v", tweet, resp, err)

	return true
}
