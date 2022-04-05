// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package pooling // import "github.com/wabarc/wayback/pooling"

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
)

func TestRoll(t *testing.T) {
	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	logger.SetLogLevel(logger.LevelFatal)

	c := 2
	p := New(context.Background(), c)
	p.timeout = 10 * time.Millisecond

	if l := len(p.resource); l != c {
		t.Fatalf("The length of pool got %d instead of %d", l, c)
	}

	capacity := 10
	var i int
	for i < capacity {
		ch := make(chan struct{}, 1)
		go func(i int) {
			bucket := &Bucket{
				Request: func(_ context.Context) error {
					time.Sleep(time.Millisecond)
					return nil
				},
				Fallback: func(_ context.Context) error {
					return nil
				},
			}
			p.Put(bucket)
			ch <- struct{}{}
		}(i)
		i++
		<-ch
	}
	time.AfterFunc(time.Second, func() {
		p.Close()
	})
	p.Roll()

	if l := len(p.resource); l != c {
		t.Fatalf("The length of pool got %d instead of %d", l, c)
	}
}

func TestTimeout(t *testing.T) {
	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	logger.SetLogLevel(logger.LevelFatal)

	c := 2
	p := New(context.Background(), c)
	p.timeout = time.Millisecond

	if l := len(p.resource); l != c {
		t.Fatalf("The length of pool got %d instead of %d", l, c)
	}

	capacity := 10
	var i int
	for i < capacity {
		ch := make(chan struct{}, 1)
		go func(i int) {
			bucket := &Bucket{
				Request: func(_ context.Context) error {
					time.Sleep(10 * time.Millisecond)
					return nil
				},
				Fallback: func(_ context.Context) error {
					return nil
				},
			}
			p.Put(bucket)
			ch <- struct{}{}
		}(i)
		i++
		<-ch
	}
	time.AfterFunc(time.Second, func() {
		p.Close()
	})
	p.Roll()

	if l := len(p.resource); l != c {
		t.Fatalf("The length of pool got %d instead of %d", l, c)
	}
}

func TestMaxRetries(t *testing.T) {
	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	logger.SetLogLevel(logger.LevelFatal)

	bucket := &Bucket{
		Request: func(_ context.Context) error {
			time.Sleep(100 * time.Microsecond)
			return nil
		},
		Fallback: func(_ context.Context) error {
			return nil
		},
	}

	maxRetries := uint64(3)
	p := New(context.Background(), 1)
	p.timeout = time.Microsecond
	p.maxRetries = maxRetries
	go p.Roll()
	p.Put(bucket)
	p.Close()
	if bucket.elapsed-1 != maxRetries {
		t.Fatalf("Unexpected max retries got %d instead of %d", bucket.elapsed, maxRetries)
	}
}

func TestFallback(t *testing.T) {
	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	logger.SetLogLevel(logger.LevelFatal)

	want := "foo"
	fall := ""
	bucket := &Bucket{
		Request: func(_ context.Context) error {
			time.Sleep(100 * time.Microsecond)
			return errors.New("some error")
		},
		Fallback: func(_ context.Context) error {
			fall = want
			return nil
		},
	}

	p := New(context.Background(), 1)
	p.timeout = time.Microsecond
	p.maxRetries = 1
	go p.Roll()
	p.Put(bucket)
	p.Close()

	if fall != want {
		t.Fatalf("Unexpected fallback got %s instead of %s", fall, want)
	}
}

func TestClose(t *testing.T) {
	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	logger.SetLogLevel(logger.LevelFatal)

	p := New(context.Background(), 1)
	p.timeout = time.Microsecond
	go p.Roll()

	if p.resource == nil {
		t.Fatalf("Unexpected pooling resource, got <nil>")
	}
	p.Close()
	_, ok := (<-p.resource)
	if !ok {
		t.Fatalf("Unexpected close pooling")
	}
}
