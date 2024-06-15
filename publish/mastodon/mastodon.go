// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package mastodon // import "github.com/wabarc/wayback/publish/mastodon"

import (
	"context"
	"net/http"

	"github.com/mattn/go-mastodon"
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
var _ publish.Publisher = (*Mastodon)(nil)

type Mastodon struct {
	client *mastodon.Client
	opts   *config.Options
}

// New returns a Mastodon client.
func New(httpClient http.Client, opts *config.Options) *Mastodon {
	if !opts.PublishToMastodon() {
		logger.Debug("Missing required environment variable")
		return nil
	}

	client := mastodon.NewClient(&mastodon.Config{
		Server:       opts.MastodonServer(),
		ClientID:     opts.MastodonClientKey(),
		ClientSecret: opts.MastodonClientSecret(),
		AccessToken:  opts.MastodonAccessToken(),
	})
	client.Client = httpClient

	return &Mastodon{client: client, opts: opts}
}

// Publish publish toot to the Mastodon of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (m *Mastodon) Publish(ctx context.Context, rdx reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	var id string
	if len(args) > 0 {
		id = args[0]
	}
	metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusFailure)
		return errors.New("publish to mastodon: collects empty")
	}

	_, err := publish.Artifact(ctx, rdx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	var txt = render.ForPublish(&render.Mastodon{Cols: cols, Data: rdx}).String()
	if m.toMastodon(ctx, txt, id) {
		metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusSuccess)
		return nil
	}
	metrics.IncrementPublish(metrics.PublishMstdn, metrics.StatusFailure)
	return errors.New("publish to mastodon failed")
}

func (m *Mastodon) toMastodon(ctx context.Context, text, id string) bool {
	if !m.opts.PublishToMastodon() || m.client == nil {
		logger.Warn("Do not publish to Mastodon.")
		return false
	}
	if text == "" {
		logger.Warn("mastodon validation failed: Text can't be blank")
		return false
	}

	toot := &mastodon.Toot{
		Status:     text,
		SpoilerText: m.opts.MastodonCWText(),
		Visibility: mastodon.VisibilityPublic,
	}
	if id != "" {
		toot.InReplyToID = mastodon.ID(id)
	}
	if _, err := m.client.PostStatus(ctx, toot); err != nil {
		logger.Error("post Mastodon status failed: %v", err)
		return false
	}

	return true
}

// Shutdown shuts down the Mastodon publish service, it always return a nil error.
func (m *Mastodon) Shutdown() error {
	return nil
}
