// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"

	mstdn "github.com/mattn/go-mastodon"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"
)

type mastodon struct {
	client *mstdn.Client
}

// NewMastodon returns a mastodon client.
func NewMastodon(client *mstdn.Client) *mastodon {
	if !config.Opts.PublishToMastodon() {
		logger.Error("Missing required environment variable")
		return new(mastodon)
	}

	if client == nil {
		client = mstdn.NewClient(&mstdn.Config{
			Server:       config.Opts.MastodonServer(),
			ClientID:     config.Opts.MastodonClientKey(),
			ClientSecret: config.Opts.MastodonClientSecret(),
			AccessToken:  config.Opts.MastodonAccessToken(),
		})
	}

	return &mastodon{client: client}
}

// Publish publish toot to the Mastodon of given cols and args.
// A context should contain a `reduxer.Bundle` via `publish.PubBundle` constant.
func (m *mastodon) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	var id string
	if len(args) > 1 {
		id = args[1]
	}
	metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Warn("collects empty")
		return
	}

	var bnd = bundle(ctx, cols)
	var txt = render.ForPublish(&render.Mastodon{Cols: cols, Data: bnd}).String()
	if m.ToMastodon(ctx, bnd, txt, id) {
		metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusFailure)
	return
}

func (m *mastodon) ToMastodon(ctx context.Context, bundle *reduxer.Bundle, text, id string) bool {
	if !config.Opts.PublishToMastodon() || m.client == nil {
		logger.Warn("Do not publish to Mastodon.")
		return false
	}
	if text == "" {
		logger.Warn("mastodon validation failed: Text can't be blank")
		return false
	}

	toot := &mstdn.Toot{
		Status:     text,
		Visibility: mstdn.VisibilityPublic,
	}
	if id != "" {
		toot.InReplyToID = mstdn.ID(id)
	}
	if _, err := m.client.PostStatus(ctx, toot); err != nil {
		logger.Error("post Mastodon status failed: %v", err)
		return false
	}

	return true
}
