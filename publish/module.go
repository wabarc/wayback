// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"fmt"
	"sync"

	"github.com/wabarc/wayback/config"
)

var (
	publishers = make(map[Flag]Publisher)
	modules    = make(map[Flag]SetupFunc)
	mu         sync.RWMutex
)

// SetupFunc is a function type that takes a pointer
// to a config.Options struct and returns a pointer
// to a Module struct.
type SetupFunc func(*config.Options) *Module

// Module is a struct embeds the Publisher interface
// and holds a pointer to config.Options.
type Module struct {
	Publisher

	Opts *config.Options
	Flag Flag
}

// Register registers a publish client's setup function
// and allows it to be called.
func Register(flag Flag, action SetupFunc) {
	if _, exists := modules[flag]; exists {
		panic(fmt.Sprintf("module %s registered", flag))
	}

	mu.Lock()
	modules[flag] = action
	mu.Unlock()
}

func parseModule(opts *config.Options) {
	for flag, setup := range modules {
		handler := setup(opts)
		if handler != nil {
			handler.Opts = opts
			handler.Flag = flag
			publishers[flag] = handler
		}
	}
}

func loadPublisher(flag Flag) (*Module, error) {
	mu.RLock()
	defer mu.RUnlock()

	p, ok := publishers[flag].(*Module)
	if !ok {
		return nil, fmt.Errorf("publisher %s not exists", flag)
	}
	return p, nil
}
