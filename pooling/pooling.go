// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package pooling // import "github.com/wabarc/wayback/pooling"

import (
	"sync"
	"time"

	"github.com/phf/go-queue/queue"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/errors"
)

var maxTime = 5 * time.Minute

var (
	ErrPoolNotExist = errors.New("pool not exist")  // ErrPoolNotExist pool not exist
	ErrTimeout      = errors.New("process timeout") // ErrTimeout process timeout
)

var q = queue.New()

type resource struct {
	id int
}

// Pool handles a pool of services.
type Pool chan *resource

func newResource(id int) *resource {
	return &resource{id: id}
}

// New a resource pool of the specified size
// Resources are created concurrently to save resource initialization time
func New(size int) Pool {
	p := make(Pool, size)
	wg := new(sync.WaitGroup)
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func(resId int) {
			p <- newResource(resId)
			wg.Done()
		}(i)
	}
	wg.Wait()

	return p
}

// Roll wrapper service as function to the resource pool.
func (p Pool) Roll(service func()) {
	do := func(wg *sync.WaitGroup) {
		defer wg.Done()
		fn, ok := q.PopBack().(func())
		if !ok {
			logger.Error("pop service failed")
			return
		}

		r, err := p.pull()
		if err != nil {
			logger.Error("pull resources failed: %v", err)
			return
		}

		ch := make(chan bool, 1)
		go func() {
			logger.Debug("roll service func: %#v", fn)
			fn()
			ch <- true
		}()

		select {
		case <-ch:
			logger.Info("roll service completed")
		case <-time.After(maxTime):
			logger.Warn("roll service timeout")
		}

		p.push(r)
		logger.Debug("roll service completed on #%d", r.id)
	}

	// Inserts a new value service at the front of queue q.
	q.PushFront(service)

	var wg sync.WaitGroup
	wg.Add(1)

	// TODO: retry
	go do(&wg)
	wg.Wait()
}

func (p Pool) pull() (r *resource, err error) {
	select {
	case r := <-p:
		return r, nil
	case <-time.After(maxTime):
		return nil, ErrTimeout
	}
}

func (p Pool) push(r *resource) error {
	if p == nil {
		return ErrPoolNotExist
	}
	p <- r
	return nil
}
