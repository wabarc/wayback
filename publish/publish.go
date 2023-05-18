// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"sync"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/reduxer"
)

// Flag represents a type of uint8
type Flag uint8

const (
	FlagWeb      Flag = iota // FlagWeb publish from httpd service
	FlagTelegram             // FlagTelegram publish from telegram service
	FlagTwitter              // FlagTwitter publish from twitter srvice
	FlagMastodon             // FlagMastodon publish from mastodon service
	FlagDiscord              // FlagDiscord publish from discord service
	FlagMatrix               // FlagMatrix publish from matrix service
	FlagSlack                // FlagSlack publish from slack service
	FlagNostr                // FlagSlack publish from nostr
	FlagIRC                  // FlagIRC publish from relaychat service
	FlagXMPP                 // FlagXMPP publish from XMPP service
	FlagNotion               // FlagNotion is a flag for notion publish service
	FlagGitHub               // FlagGitHub is a flag for github publish service
	FlagMeili                // FlagMeili is a flag for meilisearch publish service
)

// Publisher is the interface that wraps the basic Publish method.
//
// Publish publish message to serveral media platforms, e.g. Telegram channel, GitHub Issues, etc.
// The cols must either be a []wayback.Collect, args use for specific service.
type Publisher interface {
	Publish(context.Context, reduxer.Reduxer, []wayback.Collect, ...string) error

	// Shutdown shuts down publish services.
	Shutdown() error
}

// String returns the flag as a string.
func (f Flag) String() string {
	switch f {
	case FlagWeb:
		return "httpd"
	case FlagTelegram:
		return "telegram"
	case FlagTwitter:
		return "twiter"
	case FlagMastodon:
		return "mastodon"
	case FlagDiscord:
		return "discord"
	case FlagMatrix:
		return "matrix"
	case FlagSlack:
		return "slack"
	case FlagNostr:
		return "nostr"
	case FlagIRC:
		return "irc"
	case FlagNotion:
		return "notion"
	case FlagGitHub:
		return "github"
	case FlagMeili:
		return "meilisearch"
	default:
		return "unknown"
	}
}

// Publish handles options for publish service.
type Publish struct {
	opts *config.Options
	pool *pooling.Pool
}

// New creates a Publish struct with the given context and configuration
// options. It initializes a new pooling with the context and options, and
// parses all available modules.
//
// Returns a new Publish with the provided options and pooling.
func New(ctx context.Context, opts *config.Options) *Publish {
	// parse all modules
	parseModule(opts)

	cfg := []pooling.Option{
		pooling.Capacity(len(publishers)),
		pooling.Timeout(opts.WaybackTimeout()),
		pooling.MaxRetries(opts.WaybackMaxRetries()),
	}
	pool := pooling.New(ctx, cfg...)

	return &Publish{opts: opts, pool: pool}
}

// Start starts the publish service on the underlying pooling service. It is
// blocking and should be handled in a separate goroutine.
func (p *Publish) Start() {
	p.pool.Roll()
}

// Stop stop the Publish pooling. It waits until the pool status
// is idle and then calls Stop on the pool.
//
// Stop uses a sync.Once to ensure that Stop is only called once.
func (p *Publish) Stop() {
	exec(func(mod *Module) {
		_ = mod.Shutdown() // nolint:errcheck
	})

	var once sync.Once
	for {
		if p.pool.Status() == pooling.StatusIdle {
			once.Do(func() {
				p.pool.Close()
			})
			return
		}
	}
}

// Spread accepts calls from services that with collections and various parameters.
// It prepare all available publishers and put them into pooling.
func (p *Publish) Spread(ctx context.Context, rdx reduxer.Reduxer, cols []wayback.Collect, from Flag, args ...string) {
	v := ctx.Value(from)

	exec(func(mod *Module) {
		bucket := pooling.Bucket{
			Request: func(ctx context.Context) error {
				logger.Info("requesting publishing from [%s] to [%s]...", from, mod.Flag)
				ctx = context.WithValue(ctx, from, v)
				err := mod.Publish(ctx, rdx, cols, args...)
				if err != nil {
					logger.Error("requesting publishing from [%s] to [%s] failed: %v", from, mod.Flag, err)
				}
				return err
			},
			Fallback: func(_ context.Context) error {
				return nil
			},
		}
		p.pool.Put(bucket)
	})
}

func exec(pub func(*Module)) {
	for flag := range publishers {
		mod, err := loadPublisher(flag)
		if err != nil {
			logger.Warn("load publisher failed: %v", err)
			continue
		}
		if mod == nil {
			logger.Error("module %s is nil", flag)
			continue
		}
		pub(mod)
	}
}
