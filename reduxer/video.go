// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

//go:build !with_lux

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"context"

	"github.com/wabarc/wayback/config"
)

func (m media) viaLux(ctx context.Context, cfg *config.Options) string {
	return ""
}
