// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"context"
	"testing"

	"github.com/wabarc/wayback/config"
)

func TestRegister(t *testing.T) {
	setup := func(ctx context.Context, opts Options) (*Module, error) {
		return &Module{
			Opts:     opts,
			Servicer: nil,
			Flag:     config.ServiceHTTPd,
		}, nil
	}
	// Call Register with a valid flag and the setup function we just created
	Register(config.ServiceHTTPd, setup)

	// Call Register again with the same flag, it should panic

	defer func() {
		// Clear
		delete(modules, config.ServiceHTTPd)

		if r := recover(); r == nil {
			t.Errorf("Register should have panicked")
		}
	}()

	Register(config.ServiceHTTPd, setup)
}
