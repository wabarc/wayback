// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package httpd // import "github.com/wabarc/wayback/service/httpd"

import (
	"context"

	"github.com/wabarc/wayback/service"
)

func init() {
	service.Register(service.ServiceHTTPd, setup)
}

func setup(ctx context.Context, opts service.Options) (*service.Module, error) {
	mod, err := New(ctx, opts)

	return &service.Module{
		Servicer: mod,
		Opts:     opts,
	}, err
}
