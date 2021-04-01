// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"

	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/logger"
)

func To(ctx context.Context, opts *config.Options, col []*wayback.Collect, args ...string) {
	var from string
	if len(args) > 0 {
		from = args[0]
	}

	if opts.PublishToChannel() {
		logger.Debug("[%s] publishing to channel...", from)
		ToChannel(opts, nil, Render(col))
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
		mstdn := NewMastodon(nil, opts)
		mstdn.ToMastodon(ctx, opts, mstdn.Render(col), id)
	}
	if opts.PublishToTwitter() {
		logger.Debug("[%s] publishing to Twitter...", from)
		twitter := NewTwitter(nil, opts)
		twitter.ToTwitter(ctx, opts, twitter.Render(col))
	}
	if opts.PublishToIRCChannel() {
		logger.Debug("[%s] publishing to IRC channel...", from)
		irc := NewIRC(nil, opts)
		irc.ToChannel(ctx, opts, irc.Render(col))
	}
}
