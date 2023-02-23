// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"context"
	"net/url"

	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/reduxer"
)

const (
	MsgWaybackRetrying = "wayback timeout, retrying."
	MsgWaybackTimeout  = "wayback timeout, please try later."
)

type doFunc func(cols []wayback.Collect, rdx reduxer.Reduxer) error

// Wayback in a separate goroutine.
func Wayback(ctx context.Context, opts *config.Options, urls []*url.URL, do doFunc) error {
	var done = make(chan error, 1)
	var cols []wayback.Collect
	var rdx reduxer.Reduxer
	var err error

	go func() {
		rdx, err = reduxer.Do(ctx, opts, urls...)
		if err != nil {
			done <- errors.Wrap(err, "reduxer unexpected")
			return
		}

		cols, err = wayback.Wayback(ctx, rdx, opts, urls...)
		if err != nil {
			done <- errors.Wrap(err, "wayback failed")
			return
		}
		done <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		close(done)
		if err != nil {
			return err
		}
		defer rdx.Flush()

		// push collects to the Meilisearch
		if meili != nil {
			meili.push(cols) // nolint:errcheck
		}
		return do(cols, rdx)
	}
}
