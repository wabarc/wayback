// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package slack // import "github.com/wabarc/wayback/service/slack"

import (
	"context"
	"fmt"

	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/service"
)

func init() {
	service.Register(config.ServiceSlack, setup)
}

func setup(ctx context.Context, opts service.Options) (*service.Module, error) {
	if opts.Config.SlackEnabled() {
		mod, err := New(ctx, opts)

		return &service.Module{
			Servicer: mod,
			Opts:     opts,
		}, err
	}

	return nil, fmt.Errorf("slack service disabled")
}
