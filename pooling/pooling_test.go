// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package pooling // import "github.com/wabarc/wayback/pooling"

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
)

func TestRoll(t *testing.T) {
	defer helper.CheckTest(t)

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	logger.SetLogLevel(logger.LevelFatal)

	c := 2
	p := New(context.Background(), c)
	p.timeout = 10 * time.Millisecond
	defer helper.CheckContext(p.context, t)

	if l := len(p.resource); l != c {
		t.Fatalf("The length of pool got %d instead of %d", l, c)
	}

	capacity := 10
	var i int
	for i < capacity {
		ch := make(chan struct{}, 1)
		go func(i int) {
			bucket := Bucket{
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
	defer helper.CheckTest(t)

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
			bucket := Bucket{
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
	defer helper.CheckTest(t)

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	logger.SetLogLevel(logger.LevelFatal)

	var elapsed uint64
	bucket := Bucket{
		Request: func(_ context.Context) error {
			atomic.AddUint64(&elapsed, 1)
			return errors.New("process request failed")
		},
		Fallback: func(_ context.Context) error {
			return nil
		},
	}

	maxRetries := uint64(3)
	p := New(context.Background(), 1)
	p.timeout = time.Second
	p.maxRetries = maxRetries
	go p.Roll()
	p.Put(bucket)
	p.Close()
	if elapsed != maxRetries {
		t.Fatalf("Unexpected max retries got %d instead of %d", elapsed, maxRetries)
	}
}

func TestFallback(t *testing.T) {
	defer helper.CheckTest(t)

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	logger.SetLogLevel(logger.LevelFatal)

	want := "foo"
	fall := ""
	bucket := Bucket{
		Request: func(_ context.Context) error {
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
	defer helper.CheckTest(t)

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

func TestStatus(t *testing.T) {
	defer helper.CheckTest(t)

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	logger.SetLogLevel(logger.LevelFatal)

	tests := []struct {
		run    func(*Pool)
		name   string
		status Status
	}{
		{
			func(_ *Pool) {},
			"idle",
			StatusIdle,
		},
		{
			func(p *Pool) {
				bucket := Bucket{Request: func(_ context.Context) error { return nil }}
				p.Put(bucket)
			},
			"busy",
			StatusBusy,
		},
		{
			func(p *Pool) {
				bucket := Bucket{Request: func(_ context.Context) error { time.Sleep(time.Second); return nil }}
				p.Put(bucket)
			},
			"busy",
			StatusBusy,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := New(context.Background(), 1)
			p.timeout = time.Microsecond
			go p.Roll()
			defer p.Close()

			test.run(p)
			status := p.Status()
			if status != test.status {
				t.Errorf(`Unexpected pooling status, got %v instead of %v`, status, test.status)
			}
		})
	}
}
