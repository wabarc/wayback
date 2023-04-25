// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"testing"

	"github.com/wabarc/wayback/config"
)

func TestRegister(t *testing.T) {
	setup := func(opts *config.Options) *Module {
		return &Module{
			Opts:      opts,
			Publisher: nil,
			Flag:      FlagWeb,
		}
	}

	// Call Register with a valid flag and the setup function we just created
	Register(FlagWeb, setup)

	// Call Register again with the same flag, it should panic
	defer func() {
		// Clear
		delete(modules, FlagWeb)

		if r := recover(); r == nil {
			t.Errorf("Register should have panicked")
		}
	}()
	Register(FlagWeb, setup)
}
