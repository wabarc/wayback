// Copyright 2024 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package omnivore // import "github.com/wabarc/wayback/publish/omnivore"

import (
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/publish"
)

func init() {
	publish.Register(publish.FlagOmnivore, setup)
}

func setup(opts *config.Options) *publish.Module {
	if opts.EnabledOmnivore() {
		publisher := New(nil, opts)

		return &publish.Module{
			Publisher: publisher,
			Opts:      opts,
		}
	}

	return nil
}
