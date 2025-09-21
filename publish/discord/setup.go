// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package discord // import "github.com/wabarc/wayback/publish/discord"

import (
	"context"

	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/publish"
)

func init() {
	publish.Register(publish.FlagDiscord, setup)
}

func setup(ctx context.Context, opts *config.Options) *publish.Module {
	if opts.PublishToDiscordChannel() {
		publisher := New(ctx, nil, opts)

		return &publish.Module{
			Publisher: publisher,
			Opts:      opts,
		}
	}

	return nil
}
