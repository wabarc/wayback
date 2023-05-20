// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"context"
	"fmt"
	"net/url"

	"github.com/gookit/color"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/reduxer"
)

const (
	CommandHelp     = "help"
	CommandMetrics  = "metrics"
	CommandPlayback = "playback"

	MsgWaybackRetrying = "wayback timeout, retrying."
	MsgWaybackTimeout  = "wayback timeout, please try later."
)

type doFunc func(cols []wayback.Collect, rdx reduxer.Reduxer) error

// Servicer is the interface that wraps Serve and Shutdown method.
//
// Servicer serve serveral media platforms, e.g. Telegram, Discord, etc.
type Servicer interface {
	// Serve serve a service.
	Serve() error

	// Shutdown shuts down service.
	Shutdown() error
}

// Serve runs service in a separate goroutine.
func Serve(ctx context.Context, opts Options) (errs error) {
	// parse all modules
	parseModule(ctx, opts)

	for flag := range services {
		mod, err := loadServicer(flag)
		if err != nil {
			return errors.Wrap(err, "load service failed")
		}
		if mod == nil {
			return errors.New("module not found")
		}
		go func(mod *Module) {
			logger.Info("starting %s service...", mod.Flag)
			if err := mod.Serve(); err != nil {
				errs = errors.Wrap(errs, fmt.Sprint(err))
			}
		}(mod)
	}

	return
}

// Shutdown shuts down all services.
func Shutdown() (errs error) {
	for flag := range services {
		mod, err := loadServicer(flag)
		if err != nil {
			return errors.Wrap(err, "load service failed")
		}
		if mod == nil {
			return errors.New("module not found")
		}

		logger.Info("stopping %s service...", color.Blue.Sprint(flag))
		if err = mod.Shutdown(); err != nil {
			errs = fmt.Errorf("shutdown %s failed: %w", color.Red.Sprint(flag), err)
		}
		logger.Info("stopped %s service", color.Cyan.Sprint(flag))
	}

	return
}

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
		// Keep reduxer for publish
		// defer rdx.Flush()

		return do(cols, rdx)
	}
}
