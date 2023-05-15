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
	"github.com/wabarc/wayback/ingress"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/systemd"

	_ "github.com/wabarc/wayback/ingress/register"
)

// Create channel to listen for signals.
var signalChan chan (os.Signal) = make(chan os.Signal, 1)

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
	defer cancel()

	pool := pooling.New(ctx, cfg...)
	go pool.Roll()

	// Ingress initialize
	ingress.Init(opts)

	pub := publish.New(ctx, opts)
	go pub.Start()

	opt := []service.Option{
		service.Config(opts),
		service.Storage(store),
		service.Pool(pool),
		service.Publish(pub),
	}
	options := service.ParseOptions(opt...)

	err = service.Serve(ctx, options)
	if err != nil {
		logger.Error("server failed: %v", err)
	}

	if systemd.HasNotifySocket() {
		logger.Info("sending readiness notification to Systemd")

		if err := systemd.SdNotify(systemd.SdNotifyReady); err != nil {
			logger.Error("unable to send readiness notification to systemd: %v", err)
		}
	}

	handle(pool, pub, cancel)

	// Block until services closed
	<-ctx.Done()

	logger.Info("wayback service stopped.")
}

func handle(pool *pooling.Pool, pub *publish.Publish, cancel context.CancelFunc) {
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
	service.Shutdown() // nolint:errcheck
	// Gracefully closes the worker pool
	pool.Close()
	// Stop publish service
	pub.Stop()

	cancel()
}
