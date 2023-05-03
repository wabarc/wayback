// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"context"
	"testing"
)

func TestRegister(t *testing.T) {
	setup := func(ctx context.Context, opts Options) (*Module, error) {
		return &Module{
			Opts:     opts,
			Servicer: nil,
			Flag:     ServiceHTTPd,
		}, nil
	}
	// Call Register with a valid flag and the setup function we just created
	Register(ServiceHTTPd, setup)

	// Call Register again with the same flag, it should panic

	defer func() {
		// Clear
		delete(modules, ServiceHTTPd)

		if r := recover(); r == nil {
			t.Errorf("Register should have panicked")
		}
	}()

	Register(ServiceHTTPd, setup)
}
