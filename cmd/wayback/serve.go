// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/service/discord"
	"github.com/wabarc/wayback/service/httpd"
	"github.com/wabarc/wayback/service/mastodon"
	"github.com/wabarc/wayback/service/matrix"
	"github.com/wabarc/wayback/service/relaychat"
	"github.com/wabarc/wayback/service/slack"
	"github.com/wabarc/wayback/service/telegram"
	"github.com/wabarc/wayback/service/twitter"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/systemd"
)

// Create channel to listen for signals.
var signalChan chan (os.Signal) = make(chan os.Signal, 1)

type target struct {
	call func()
	name string
}

type services struct {
	targets []target
}

func serve(_ *cobra.Command, opts *config.Options, _ []string) {
	store, err := storage.Open(opts, "")
	if err != nil {
		logger.Fatal("open storage failed: %v", err)
	}
	defer store.Close()

	ctx, cancel := context.WithCancel(context.Background())
	pool := pooling.New(ctx, opts)
	go pool.Roll()

	if opts.EnabledMeilisearch() {
		endpoint := opts.WaybackMeiliEndpoint()
		indexing := opts.WaybackMeiliIndexing()
		apikey := opts.WaybackMeiliApikey()
		meili := service.NewMeili(endpoint, apikey, indexing)
		if err := meili.Setup(); err != nil {
			logger.Error("setup meilisearch failed: %v", err)
		}
		logger.Debug("setup meilisearch success")
	}

	srv := &services{}
	_ = srv.run(ctx, store, opts, pool)

	if systemd.HasNotifySocket() {
		logger.Info("sending readiness notification to Systemd")

		if err := systemd.SdNotify(systemd.SdNotifyReady); err != nil {
			logger.Error("unable to send readiness notification to systemd: %v", err)
		}
	}

	go srv.daemon(pool, cancel)
	<-ctx.Done()

	logger.Info("wayback service stopped.")
}

// nolint:gocyclo
func (srv *services) run(ctx context.Context, store *storage.Storage, opts *config.Options, pool *pooling.Pool) *services {
	size := len(daemon)
	srv.targets = make([]target, 0, size)
	for _, s := range daemon {
		switch s {
		case "irc":
			irc := relaychat.New(ctx, store, opts, pool)
			go func() {
				if err := irc.Serve(); err != relaychat.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { irc.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "slack":
			sl := slack.New(ctx, store, opts, pool)
			go func() {
				if err := sl.Serve(); err != slack.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { sl.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "discord":
			d := discord.New(ctx, store, opts, pool)
			go func() {
				if err := d.Serve(); err != discord.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { d.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "mastodon", "mstdn":
			m := mastodon.New(ctx, store, opts, pool)
			go func() {
				if err := m.Serve(); err != mastodon.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { m.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "telegram":
			t := telegram.New(ctx, store, opts, pool)
			go func() {
				if err := t.Serve(); err != telegram.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { t.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "twitter":
			t := twitter.New(ctx, store, opts, pool)
			go func() {
				if err := t.Serve(); err != twitter.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { t.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "matrix":
			m := matrix.New(ctx, store, opts, pool)
			go func() {
				if err := m.Serve(); err != matrix.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { m.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "web", "httpd":
			h := httpd.New(ctx, store, opts, pool)
			go func() {
				if err := h.Serve(); err != httpd.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { h.Shutdown() }, // nolint:errcheck
				name: s,
			})
		default:
			logger.Error("unrecognize %s in `--daemon`", s)
		}
	}

	return srv
}

func (srv *services) daemon(pool *pooling.Pool, cancel context.CancelFunc) {
	// SIGINT handles Ctrl+C locally.
	// SIGTERM handles termination signal from cloud service.
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		os.Interrupt,
	)

	// Receive output from signalChan.
	sig := <-signalChan
	logger.Info("signal %s is received, exiting...", sig)

	// Gracefully shutdown the server
	srv.shutdown()
	// Gracefully closes the worker pool
	pool.Close()
	cancel()
}

func (srv *services) shutdown() {
	for _, target := range srv.targets {
		logger.Info("stopping %s service...", target.name)
		target.call()
		logger.Info("stopped %s service.", target.name)
	}
}
