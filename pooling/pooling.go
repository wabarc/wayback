// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package pooling // import "github.com/wabarc/wayback/pooling"

import (
	"sync"
	"time"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/errors"
)

const maxTime = 5 * time.Minute

var (
	ErrPoolNotExist = errors.New("pool not exist")
	ErrTimeout      = errors.New("process timeout")
)

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
	do := func(service func(), wg *sync.WaitGroup) {
		defer wg.Done()
		r, err := p.pull()
		if err != nil {
			logger.Error("[pooling] pull resources failed: %v", err)
			return
		}
		logger.Debug("[pooling] roll service on #%d", r.id)
		defer p.push(r)

		logger.Debug("[pooling] roll service func: %#v", service)
		service()
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go do(service, &wg)
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
