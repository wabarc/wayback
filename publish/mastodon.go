// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/template/render"

	mstdn "github.com/mattn/go-mastodon"
)

var _ Publisher = (*mastodon)(nil)

type mastodon struct {
	client *mstdn.Client
	opts   *config.Options
}

// NewMastodon returns a mastodon client.
func NewMastodon(client *mstdn.Client, opts *config.Options) *mastodon {
	if !opts.PublishToMastodon() {
		logger.Error("Missing required environment variable")
		return new(mastodon)
	}

	if client == nil {
		client = mstdn.NewClient(&mstdn.Config{
			Server:       opts.MastodonServer(),
			ClientID:     opts.MastodonClientKey(),
			ClientSecret: opts.MastodonClientSecret(),
			AccessToken:  opts.MastodonAccessToken(),
		})
	}

	return &mastodon{client: client, opts: opts}
}

// Publish publish toot to the Mastodon of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (m *mastodon) Publish(ctx context.Context, cols []wayback.Collect, args ...string) error {
	var id string
	if len(args) > 1 {
		id = args[1]
	}
	metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusRequest)

	if len(cols) == 0 {
		return errors.New("publish to mastodon: collects empty")
	}

	rdx, _, err := extract(ctx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	var txt = render.ForPublish(&render.Mastodon{Cols: cols, Data: rdx}).String()
	if m.ToMastodon(ctx, txt, id) {
		metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusSuccess)
		return nil
	}
	metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusFailure)
	return errors.New("publish to mastodon failed")
}

func (m *mastodon) ToMastodon(ctx context.Context, text, id string) bool {
	if !m.opts.PublishToMastodon() || m.client == nil {
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
