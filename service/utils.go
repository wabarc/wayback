// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

// MatchURL returns a slice string contains URLs extracted from the given string.
func MatchURL(s string) []string {
	if config.Opts.WaybackFallback() {
		return helper.MatchURLFallback(s)
	}
	return helper.MatchURL(s)
}
