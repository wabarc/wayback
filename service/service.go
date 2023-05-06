// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"context"
	"fmt"
	"net/url"

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

const (
	ServiceDiscord  Flag = iota + 1 // FlagDiscord represents discord service
	ServiceHTTPd                    // FlagWeb represents httpd service
	ServiceMastodon                 // FlagMastodon represents mastodon service
	ServiceMatrix                   // FlagMatrix represents matrix service
	ServiceIRC                      // FlagIRC represents relaychat service
	ServiceSlack                    // FlagSlack represents slack service
	ServiceTelegram                 // FlagTelegram represents telegram service
	ServiceTwitter                  // FlagTwitter represents twitter srvice
	ServiceXMPP                     // FlagXMPP represents XMPP service
)

// Flag represents a type of uint8
type Flag uint8

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

// String returns the flag as a string.
func (f Flag) String() string {
	switch f {
	case ServiceHTTPd:
		return "httpd"
	case ServiceTelegram:
		return "telegram"
	case ServiceTwitter:
		return "twiter"
	case ServiceMastodon:
		return "mastodon"
	case ServiceDiscord:
		return "discord"
	case ServiceMatrix:
		return "matrix"
	case ServiceSlack:
		return "slack"
	case ServiceIRC:
		return "relaychat"
	case ServiceXMPP:
		return "xmpp"
	default:
		return ""
	}
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

		logger.Info("stopping %s service...", flag)
		if err = mod.Shutdown(); err != nil {
			errs = fmt.Errorf("shutdown %s failed: %w", flag, err)
		}
		logger.Info("stopped %s service", flag)
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
