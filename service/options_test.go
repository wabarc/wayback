// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"testing"

	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/storage"
)

func TestParseOptions(t *testing.T) {
	configOpts := &config.Options{}
	pool := &pooling.Pool{}
	publish := &publish.Publish{}
	storage := &storage.Storage{}

	opts := ParseOptions(
		Config(configOpts),
		Pool(pool),
		Publish(publish),
		Storage(storage),
	)

	if opts.Config != configOpts {
		t.Errorf("Unexpected config options: expected=%v, got=%v", configOpts, opts.Config)
	}

	if opts.Pool != pool {
		t.Errorf("Unexpected pooling options: expected=%v, got=%v", pool, opts.Pool)
	}

	if opts.Publish != publish {
		t.Errorf("Unexpected publish options: expected=%v, got=%v", publish, opts.Publish)
	}

	if opts.Storage != storage {
		t.Errorf("Unexpected storage options: expected=%v, got=%v", storage, opts.Storage)
	}
}

func TestConfig(t *testing.T) {
	configOpts := &config.Options{}
	option := Config(configOpts)

	opts := &Options{}
	option(opts)

	if opts.Config != configOpts {
		t.Errorf("Unexpected config options: expected=%v, got=%v", configOpts, opts.Config)
	}
}

func TestPool(t *testing.T) {
	pool := &pooling.Pool{}
	option := Pool(pool)

	opts := &Options{}
	option(opts)

	if opts.Pool != pool {
		t.Errorf("Unexpected pooling options: expected=%v, got=%v", pool, opts.Pool)
	}
}

func TestPublish(t *testing.T) {
	publish := &publish.Publish{}
	option := Publish(publish)

	opts := &Options{}
	option(opts)

	if opts.Publish != publish {
		t.Errorf("Unexpected publish options: expected=%v, got=%v", publish, opts.Publish)
	}
}

func TestStorage(t *testing.T) {
	storage := &storage.Storage{}
	option := Storage(storage)

	opts := &Options{}
	option(opts)

	if opts.Storage != storage {
		t.Errorf("Unexpected storage options: expected=%v, got=%v", storage, opts.Storage)
	}
}
