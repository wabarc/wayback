// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package ingress // import "github.com/wabarc/wayback/ingress"

import (
	"github.com/wabarc/wayback/config"
)

// Init initializes functions with the given configuration options.
func Init(opts *config.Options) {
	initClient(opts)
}
