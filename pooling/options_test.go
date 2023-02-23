// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package pooling // import "github.com/wabarc/wayback/pooling"

import (
	"testing"
	"time"
)

func TestTimeoutOption(t *testing.T) {
	opts := &Options{}

	timeout := time.Second
	Timeout(timeout)(opts)

	if opts.Timeout != timeout {
		t.Errorf("Expected timeout to be seconds, but got %v", opts.Timeout)
	}
}

func TestMaxRetriesOption(t *testing.T) {
	opts := &Options{}

	MaxRetries(3)(opts)

	if opts.MaxRetries != 3 {
		t.Errorf("Expected max retries to be 3, but got %v", opts.MaxRetries)
	}
}

func TestCapacityOption(t *testing.T) {
	opts := &Options{}

	Capacity(100)(opts)

	if opts.Capacity != 100 {
		t.Errorf("Expected capacity to be 100, but got %v", opts.Capacity)
	}
}
