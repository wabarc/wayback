// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package pooling // import "github.com/wabarc/wayback/pooling"

import (
	"time"
)

// Options represents configuration for pooling.
type Options struct {
	Timeout    time.Duration // Timeout specifies the maximum amount of time to wait for an operation to complete.
	MaxRetries uint64        // MaxRetries specifies the maximum number of times to retry the operation in case of failure.
	Capacity   int           // Capacity specifies the maximum number of items that can be processed simultaneously.
}

// Option is a function that modifies the provided Options instance.
type Option func(*Options)

// Timeout returns an Option function that sets the timeout value for the Options.
// The given timeout value will be applied when the returned function is called.
func Timeout(t time.Duration) Option {
	return func(opts *Options) {
		opts.Timeout = t
	}
}

// MaxRetries returns an Option function that sets the maximum retries value for the Options.
// The given maximum retries value will be applied when the returned function is called.
func MaxRetries(r uint64) Option {
	return func(opts *Options) {
		opts.MaxRetries = r
	}
}

// Capacity returns an Option function that sets the capacity value for the Options.
// The given capacity value will be applied when the returned function is called.
func Capacity(c int) Option {
	return func(opts *Options) {
		opts.Capacity = c
	}
}
