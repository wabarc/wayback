// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

//go:build !with_tor

package httpd // import "github.com/wabarc/wayback/service/httpd"

import (
	"github.com/cretz/bine/process"
)

var creator process.Creator
