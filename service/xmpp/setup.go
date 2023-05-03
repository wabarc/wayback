// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package xmpp // import "github.com/wabarc/wayback/service/xmpp"

import (
	"context"
	"fmt"

	"github.com/wabarc/wayback/service"
)

func init() {
	service.Register(service.ServiceXMPP, setup)
}

func setup(ctx context.Context, opts service.Options) (*service.Module, error) {
	if opts.Config.XMPPEnabled() {
		mod, err := New(ctx, opts)

		return &service.Module{
			Servicer: mod,
			Opts:     opts,
		}, err
	}

	return nil, fmt.Errorf("xmpp service disabled")
}
