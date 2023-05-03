// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/service/telegram"

import (
	"context"
	"fmt"

	"github.com/wabarc/wayback/service"
)

func init() {
	service.Register(service.ServiceTelegram, setup)
}

func setup(ctx context.Context, opts service.Options) (*service.Module, error) {
	if opts.Config.TelegramEnabled() {
		mod, err := New(ctx, opts)

		return &service.Module{
			Servicer: mod,
			Opts:     opts,
		}, err
	}

	return nil, fmt.Errorf("telegram service disabled")
}
