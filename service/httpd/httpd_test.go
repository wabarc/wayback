// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package httpd // import "github.com/wabarc/wayback/service/httpd"

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"
)

func TestNew(t *testing.T) {
	opts := service.Options{
		Config:  &config.Options{},
		Pool:    &pooling.Pool{},
		Storage: &storage.Storage{},
		Publish: &publish.Publish{},
	}

	tests := []struct {
		ctx  context.Context
		opts service.Options
		name string
	}{
		{
			ctx:  context.Background(),
			opts: opts,
			name: "ok",
		},
		{
			ctx:  nil,
			opts: opts,
			name: "nil context",
		},
		{
			ctx:  context.Background(),
			opts: service.Options{},
			name: "nil options",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := test.ctx
			opts := test.opts

			h := New(ctx, opts)
			if h.opts != opts.Config {
				t.Errorf("Expected config options to be %v, but got %v", opts.Config, h.opts)
			}
			if h.pool != opts.Pool {
				t.Errorf("Expected pool to be %v, but got %v", opts.Pool, h.pool)
			}
			if h.store != opts.Storage {
				t.Errorf("Expected storage to be %v, but got %v", opts.Storage, h.store)
			}
			if h.pub != opts.Publish {
				t.Errorf("Expected publish to be %v, but got %v", opts.Publish, h.pub)
			}
		})
	}
}

func TestHttpdServe(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	min := 1000
	max := 65535
	port := strconv.Itoa(rand.Intn(max-min+1) + min)
	addr := "127.0.0.1"

	t.Setenv("WAYBACK_LISTEN_ADDR", net.JoinHostPort(addr, port))
	opts, err := config.NewParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	// Makes tor binary not exists
	t.Setenv("PATH", "")

	ctx := context.Background()
	httpd := New(ctx, service.Options{
		Config:  opts,
		Storage: &storage.Storage{},
		Pool:    &pooling.Pool{},
		Publish: &publish.Publish{},
	})

	go func() {
		err := httpd.Serve()
		if err != ErrServiceClosed {
			t.Errorf("Expected ErrServiceClosed, got %v", err)
		}
	}()

	// Wait for the server to start before calling shutdown
	time.Sleep(100 * time.Millisecond)

	// Shutdown the server and assert that there were no errors
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := httpd.Shutdown()
		if err != nil {
			t.Errorf("Expected no error during shutdown, got %v", err)
		}
	}()

	// Wait for both methods to finish
	wg.Wait()
}
