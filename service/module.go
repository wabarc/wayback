// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"context"
	"fmt"
	"sync"
)

var (
	services = make(map[Flag]Servicer)
	modules  = make(map[Flag]SetupFunc)
	mu       sync.RWMutex
)

// SetupFunc is a function type that takes a pointer
// to a Options struct and returns a pointer
// to a Module struct.
type SetupFunc func(context.Context, Options) (*Module, error)

// Module is a struct embeds the Servicer interface
// and holds a pointer to Options.
type Module struct {
	Servicer

	Opts Options
	Flag Flag
}

// Register registers a service instance's setup function
// and allows it to be called.
func Register(srv Flag, action SetupFunc) {
	if _, exists := modules[srv]; exists {
		panic(fmt.Sprintf("module %s registered", srv))
	}

	mu.Lock()
	modules[srv] = action
	mu.Unlock()
}

func parseModule(ctx context.Context, opts Options) {
	for flag, setup := range modules {
		mod, err := setup(ctx, opts)
		if err == nil && mod != nil {
			mod.Opts = opts
			mod.Flag = flag
			services[flag] = mod
		}
	}
}

func loadServicer(flag Flag) (*Module, error) {
	mu.RLock()
	defer mu.RUnlock()

	p, ok := services[flag].(*Module)
	if !ok {
		return nil, fmt.Errorf("servicer %s not exists", flag)
	}
	return p, nil
}
