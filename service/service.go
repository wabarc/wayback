// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"context"
	"net/url"
	"sync"

	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/reduxer"
)

const (
	MsgWaybackRetrying = "wayback timeout, retrying."
	MsgWaybackTimeout  = "wayback timeout, please try later."
)

// Wayback in a separate goroutine.
func Wayback(ctx context.Context, urls []*url.URL, do func(cols []wayback.Collect, rdx reduxer.Reduxer) error) error {
	var once sync.Once
	var done = make(chan bool, 1)
	var cols []wayback.Collect
	var rdx reduxer.Reduxer
	var err error
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			err = do(cols, rdx)
			rdx.Flush()
			return err
		default:
			once.Do(func() {
				cols, rdx, err = wayback.Wayback(ctx, urls...)
				if err != nil {
					err = errors.Wrap(err, "wayback failed")
				} else {
					done <- true
				}
			})
		}
	}
}
