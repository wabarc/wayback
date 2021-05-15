// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package storage // import "github.com/wabarc/wayback/storage"

import (
	"os"
	"path/filepath"

	"github.com/wabarc/helper"
)

func tmpPath() string {
	r := helper.RandString(5, "lower")
	return filepath.Join(os.TempDir(), r)
}
