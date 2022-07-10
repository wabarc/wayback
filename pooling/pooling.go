// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package pooling // import "github.com/wabarc/wayback/pooling"

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/phf/go-queue/queue"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
)

var (
	ErrPoolNotExist = errors.New("pool not exist")  // ErrPoolNotExist pool not exist
	ErrTimeout      = errors.New("process timeout") // ErrTimeout process timeout

	errRollTimeout = errors.New("roll bucket timeout")
	errElapsed     = errors.New("retried to reach maximum times")
)

type resource struct {
	id int
}

// Pool represents a pool of services.
type Pool struct {
	mutex    sync.Mutex
	resource chan *resource
	timeout  time.Duration
	staging  queue.Queue

	closed  chan bool
	context context.Context

	waiting    int32
	processing int32
	maxRetries uint64
	multiplier float64
}

// A Bucket represents a wayback request is sent by a service.
type Bucket struct {
	// Request is the main func for handling wayback requests.
	Request func(context.Context) error

	// Fallback defines an optional func to return a failure response for the Request func.
	Fallback func(context.Context) error

	// Count of retried attempts
	elapsed uint64

	// An object that will perform exactly one action.
	once *sync.Once
}

func newResource(id int) *resource {
	return &resource{id: id}
}

// New a resource pool of the specified capacity
// Resources are created concurrently to save resource initialization time
func New(ctx context.Context, capacity int) *Pool {
	p := new(Pool)
	p.resource = make(chan *resource, capacity)
	wg := new(sync.WaitGroup)
	wg.Add(capacity)
	for i := 0; i < capacity; i++ {
		go func(resId int) {
			p.resource <- newResource(resId)
			wg.Done()
		}(i)
	}
	wg.Wait()

	p.closed = make(chan bool, 1)
	p.timeout = config.Opts.WaybackTimeout()
	p.maxRetries = config.Opts.WaybackMaxRetries() + 1
	p.multiplier = 0.75
	p.context = ctx

	return p
}

// Roll process wayback requests from the resource pool for execution.
//
//  // Stream generates values with DoSomething and sends them to out
//  // until DoSomething returns an error or ctx.Done is closed.
//  func Stream(ctx context.Context, out chan<- Value) error {
//  	for {
//  		v, err := DoSomething(ctx)
//  		if err != nil {
//  			return err
//  		}
//  		select {
//  		case <-ctx.Done():
//  			return ctx.Err()
//  		case out <- v:
//  		}
//  	}
//  }
//
func (p *Pool) Roll() {
	// Blocks until closed
	for {
		select {
		default:
		case <-p.closed:
			close(p.closed)
			return
		}

		// Waiting for new requests
		if atomic.LoadInt32(&p.waiting) == 0 {
			continue
		}

		if b, has := p.bucket(); has {
			go b.once.Do(func() {
				p.do(b)
			})
		}
	}
}

// Pub puts wayback requests to the resource pool
func (p *Pool) Put(b Bucket) {
	// Inserts a new bucket at the front of queue.
	p.mutex.Lock()
	p.staging.PushFront(b)
	p.mutex.Unlock()
	atomic.AddInt32(&p.waiting, 1)
}

// Close closes the worker pool, and it is blocked until all workers are idle.
func (p *Pool) Close() {
	var once sync.Once
	for {
		waiting := atomic.LoadInt32(&p.waiting)
		processing := atomic.LoadInt32(&p.processing)
		if p.resource != nil && waiting == 0 && processing == 0 {
			once.Do(func() {
				p.closed <- true
			})
			return
		}
	}
}

func (p *Pool) pull() (r *resource) {
	select {
	case r = <-p.resource:
		return r
	}
}

func (p *Pool) push(r *resource) error {
	if p == nil {
		return ErrPoolNotExist
	}
	p.resource <- r
	return nil
}

func (p *Pool) do(b Bucket) error {
	atomic.AddInt32(&p.processing, 1)
	defer func() {
		atomic.AddInt32(&p.waiting, -1)
		atomic.AddInt32(&p.processing, -1)
	}()

	action := func() error {
		interval := float64(b.elapsed) * p.multiplier
		timeout := p.timeout + p.timeout*time.Duration(interval)
		ctx, cancel := context.WithTimeout(p.context, timeout)
		defer cancel()

		r := p.pull()
		defer func() {
			p.push(r)
			if b.elapsed >= p.maxRetries {
				if b.Fallback != nil {
					b.Fallback(ctx)
				}
			}
		}()

		ch := make(chan error, 1)
		go func() {
			if b.Request != nil {
				ch <- b.Request(ctx)
			}
		}()

		select {
		case err := <-ch:
			if err != nil {
				atomic.AddUint64(&b.elapsed, 1)
			}
			close(ch)
			return err
		case <-ctx.Done():
			atomic.AddUint64(&b.elapsed, 1)
			return errRollTimeout
		}
	}

	ran := uint64(1)
	max := p.maxRetries
	for ; ran <= max; ran++ {
		err := action()
		switch err {
		case nil, errElapsed:
			return err
		case errRollTimeout:
		}
	}

	return nil
}

func (p *Pool) bucket() (b Bucket, ok bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if b, ok = p.staging.PopBack().(Bucket); ok {
		b.once = new(sync.Once)
		return b, ok
	}

	return
}
