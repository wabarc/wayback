// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/service/discord"
	"github.com/wabarc/wayback/service/httpd"
	"github.com/wabarc/wayback/service/mastodon"
	"github.com/wabarc/wayback/service/matrix"
	"github.com/wabarc/wayback/service/relaychat"
	"github.com/wabarc/wayback/service/slack"
	"github.com/wabarc/wayback/service/telegram"
	"github.com/wabarc/wayback/service/twitter"
	"github.com/wabarc/wayback/service/xmpp"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/systemd"

	_ "github.com/wabarc/wayback/ingress"
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

	cfg := []pooling.Option{
		pooling.Capacity(opts.PoolingSize()),
		pooling.Timeout(opts.WaybackTimeout()),
		pooling.MaxRetries(opts.WaybackMaxRetries()),
	}
	ctx, cancel := context.WithCancel(context.Background())
	pool := pooling.New(ctx, cfg...)
	go pool.Roll()

	pub := publish.New(ctx, opts)
	go pub.Start()

	opt := []service.Option{
		service.Config(opts),
		service.Storage(store),
		service.Pool(pool),
		service.Publish(pub),
	}
	options := service.ParseOptions(opt...)

	srv := &services{}
	_ = srv.run(ctx, options)

	if systemd.HasNotifySocket() {
		logger.Info("sending readiness notification to Systemd")

		if err := systemd.SdNotify(systemd.SdNotifyReady); err != nil {
			logger.Error("unable to send readiness notification to systemd: %v", err)
		}
	}

	go srv.daemon(pool, pub, cancel)
	<-ctx.Done()

	logger.Info("wayback service stopped.")
}

// nolint:gocyclo
func (srv *services) run(ctx context.Context, opts service.Options) *services {
	size := len(daemon)
	srv.targets = make([]target, 0, size)
	for _, s := range daemon {
		s := s
		switch strings.ToLower(s) {
		case "irc":
			irc := relaychat.New(ctx, opts)
			go func() {
				if err := irc.Serve(); err != relaychat.ErrServiceClosed {
					logger.Error("start %s service failed: %v", s, err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { irc.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "slack":
			sl := slack.New(ctx, opts)
			go func() {
				if err := sl.Serve(); err != slack.ErrServiceClosed {
					logger.Error("start %s service failed: %v", s, err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { sl.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "discord":
			d := discord.New(ctx, opts)
			go func() {
				if err := d.Serve(); err != discord.ErrServiceClosed {
					logger.Error("start %s service failed: %v", s, err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { d.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "mastodon", "mstdn":
			m := mastodon.New(ctx, opts)
			go func() {
				if err := m.Serve(); err != mastodon.ErrServiceClosed {
					logger.Error("start %s service failed: %v", s, err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { m.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "telegram":
			t := telegram.New(ctx, opts)
			go func() {
				if err := t.Serve(); err != telegram.ErrServiceClosed {
					logger.Error("start %s service failed: %v", s, err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { t.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "twitter":
			t := twitter.New(ctx, opts)
			go func() {
				if err := t.Serve(); err != twitter.ErrServiceClosed {
					logger.Error("start %s service failed: %v", s, err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { t.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "matrix":
			m := matrix.New(ctx, opts)
			go func() {
				if err := m.Serve(); err != matrix.ErrServiceClosed {
					logger.Error("start %s service failed: %v", s, err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { m.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "web", "httpd":
			h := httpd.New(ctx, opts)
			go func() {
				if err := h.Serve(); err != httpd.ErrServiceClosed {
					logger.Error("start %s service failed: %v", s, err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { h.Shutdown() }, // nolint:errcheck
				name: s,
			})
		case "jabber", "xmpp":
			h := xmpp.New(ctx, opts)
			go func() {
				if err := h.Serve(); err != xmpp.ErrServiceClosed {
					logger.Error("start %s service failed: %v", s, err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { h.Shutdown() }, // nolint:errcheck
				name: s,
			})
		default:
			logger.Fatal("unrecognize %s in `--daemon`", s)
		}
	}

	return srv
}

func (srv *services) daemon(pool *pooling.Pool, pub *publish.Publish, cancel context.CancelFunc) {
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
	// Stop publish service
	pub.Stop()
	cancel()
}

func (srv *services) shutdown() {
	for _, target := range srv.targets {
		logger.Info("stopping %s service...", target.name)
		target.call()
		logger.Info("stopped %s service.", target.name)
	}
}
