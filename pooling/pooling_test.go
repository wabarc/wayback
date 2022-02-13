// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package pooling // import "github.com/wabarc/wayback/pooling"

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
)

func TestTimeout(t *testing.T) {
	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	logger.SetLogLevel(logger.LevelFatal)

	c := 2
	p := New(c)
	p.timeout = time.Microsecond

	var i int32
	var wg sync.WaitGroup
	for i < 5 {
		wg.Add(1)
		p.Roll(func() {
			time.Sleep(time.Millisecond)
		})
		wg.Done()
		atomic.AddInt32(&i, 1)
		runtime.Gosched()
	}
	wg.Wait()

	if l := len(p.resource); l != c {
		t.Fatalf("The length of pool got %d instead of %d", l, c)
	}

	p.Roll(func() {
		time.Sleep(time.Millisecond)
	})

	if l := len(p.resource); l != c {
		t.Fatalf("The length of pool got %d instead of %d", l, c)
	}
}

func BenchmarkRoll(b *testing.B) {
	p := New(1)
	for n := 0; n < b.N; n++ {
		p.Roll(func() {
			time.Sleep(time.Millisecond)
		})
	}
}
