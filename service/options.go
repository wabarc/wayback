// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/storage"
)

// Options represents the configuration for services.
type Options struct {
	// Config holds the configuration options.
	Config *config.Options

	// Pool holds the connection pool to be used.
	Pool *pooling.Pool

	// Publish holds the publish service to be used.
	Publish *publish.Publish

	// Storage holds the storage service to be used.
	Storage *storage.Storage
}

// Option is a function that modifies the provided Options instance.
type Option func(*Options)

// ParseOptions returns the Options instance with modifications applied using the provided Option functions.
func ParseOptions(opts ...Option) (o Options) {
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// Config returns an Option function that sets the Config field of Options.
func Config(c *config.Options) Option {
	return func(opts *Options) {
		opts.Config = c
	}
}

// Pool returns an Option function that sets the Pool field of Options.
func Pool(p *pooling.Pool) Option {
	return func(opts *Options) {
		opts.Pool = p
	}
}

// Publish returns an Option function that sets the Publish field of Options.
func Publish(p *publish.Publish) Option {
	return func(opts *Options) {
		opts.Publish = p
	}
}

// Storage returns an Option function that sets the Storage field of Options.
func Storage(s *storage.Storage) Option {
	return func(opts *Options) {
		opts.Storage = s
	}
}
