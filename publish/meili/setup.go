// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package meili // import "github.com/wabarc/wayback/publish/meili"

import (
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/publish"
)

func init() {
	publish.Register(publish.FlagMeili, setup)
}

func setup(opts *config.Options) *publish.Module {
	if opts.EnabledMeilisearch() {
		publisher := New(nil, opts)

		// Setup meilisearch
		err := publisher.setup()
		if err != nil {
			logger.Error("setup meilisearch failed: %v", err)
			return nil
		}

		return &publish.Module{
			Publisher: publisher,
			Opts:      opts,
		}
	}

	return nil
}
