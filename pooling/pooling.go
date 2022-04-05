// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package pooling // import "github.com/wabarc/wayback/pooling"

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
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

	// Reports whether it is process has been completed.
	// processed chan bool
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
	p.maxRetries = config.Opts.WaybackMaxRetries()
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
			return
		}

		// Waiting for new requests
		if atomic.LoadInt32(&p.waiting) == 0 {
			continue
		}

		b := p.bucket()
		if b == nil {
			continue
		}
		go p.do(b, b.Request, b.Fallback)
	}
}

// Pub puts wayback requests to the resource pool
func (p *Pool) Put(b *Bucket) {
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
				close(p.resource)
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
	p.mutex.Lock()
	p.resource <- r
	p.mutex.Unlock()
	return nil
}

func (p *Pool) do(b *Bucket, request, fallback func(context.Context) error) error {
	atomic.AddInt32(&p.processing, 1)
	defer func() {
		atomic.AddInt32(&p.waiting, -1)
		atomic.AddInt32(&p.processing, -1)
	}()

	action := func() error {
		ctx, cancel := context.WithCancel(p.context)
		defer cancel()

		if b.elapsed > p.maxRetries {
			if fallback != nil {
				fallback(ctx)
			}
			return errElapsed
		}

		r := p.pull()
		defer func() {
			p.push(r)
		}()

		res := make(chan error, 1)
		go func() {
			if request != nil {
				res <- request(ctx)
				return
			}
			res <- nil
			return
		}()

		interval := float64(b.elapsed) * p.multiplier
		timeout := p.timeout + p.timeout*time.Duration(interval)
		for {
			select {
			case err := <-res:
				if err != nil {
					atomic.AddUint64(&b.elapsed, 1)
				}
				return err
			case <-time.After(timeout):
				atomic.AddUint64(&b.elapsed, 1)
				return errRollTimeout
			}
		}
	}

	return p.doRetry(action)
}

func (p *Pool) bucket() *Bucket {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	b, ok := p.staging.PopBack().(*Bucket)
	if !ok {
		return nil
	}

	return b
}

func (p *Pool) doRetry(o backoff.Operation) error {
	exp := backoff.NewExponentialBackOff()
	exp.Reset()
	b := backoff.WithMaxRetries(exp, p.maxRetries+1) // One more retry for fallback to be triggered

	return backoff.Retry(o, b)
}
