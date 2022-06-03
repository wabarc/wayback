// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"context"
	"net/url"

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
	var done = make(chan error, 1)
	var cols []wayback.Collect
	var rdx reduxer.Reduxer

	go func() {
		go func() {
			var err error
			cols, rdx, err = wayback.Wayback(ctx, urls...)
			if err != nil {
				done <- errors.Wrap(err, "wayback failed")
				return
			}
			defer rdx.Flush()
			// push collects to the Meilisearch
			if meili != nil {
				go meili.push(cols)
			}
			done <- do(cols, rdx)
		}()

		// Block until context is finished.
		select {
		case <-ctx.Done():
			return
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		close(done)
		return err
	}
}
