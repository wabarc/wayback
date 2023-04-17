// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package httpd // import "github.com/wabarc/wayback/service/httpd"

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/fatih/color"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"
)

// ErrServiceClosed is returned by the Service's Serve method after a call to Shutdown.
var ErrServiceClosed = errors.New("httpd: Service closed")

// Httpd represents a http server in the application.
type Httpd struct {
	sync.RWMutex

	ctx   context.Context
	pub   *publish.Publish
	opts  *config.Options
	pool  *pooling.Pool
	store *storage.Storage

	tor    *tor.Tor
	server *http.Server
}

// New a Httpd struct.
func New(ctx context.Context, opts service.Options) *Httpd {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Httpd{
		ctx:   ctx,
		store: opts.Storage,
		opts:  opts.Config,
		pool:  opts.Pool,
		pub:   opts.Publish,
	}
}

// Serve accepts incoming HTTP requests over Tor network, or open
// a local port for proxy server by "WAYBACK_TOR_LOCAL_PORT" env.
// Use "WAYBACK_TOR_PRIVKEY" to keep the Tor hidden service hostname.
//
// Serve always returns an error.
func (h *Httpd) Serve() error {
	// Start tor with some defaults + elevated verbosity
	logger.Info("starting and registering onion service, please wait a bit...")

	handler := newWeb(h.ctx, h.opts, h.pool, h.pub).handle()
	server := &http.Server{
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  5 * time.Minute,
		Handler:      handler,
	}

	switch {
	case h.serveOnion():
		logger.Info("start a tor hidden server")
		err := h.startOnionService(server)
		if err != nil {
			return errors.Wrap(err, "start tor server failed")
		}
	default:
		logger.Info("start a clear web server")
		server.Addr = h.opts.ListenAddr()
		go startHTTPServer(server)
		h.Lock()
		h.server = server
		h.Unlock()
	}

	// Block until context done
	<-h.ctx.Done()

	return ErrServiceClosed
}

// Shutdown shuts down the httpd server
func (h *Httpd) Shutdown() error {
	h.RLock()
	defer h.RUnlock()

	// Close onion service.
	if h.tor != nil {
		if err := h.tor.Close(); err != nil {
			return err
		}
	}
	// Shutdown http server
	if h.server != nil {
		if err := h.server.Shutdown(h.ctx); err != nil {
			return err
		}
	}
	return nil
}

func startHTTPServer(server *http.Server) {
	logger.Info(`Listening on "%s" without TLS`, color.BlueString(server.Addr))
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal("Server failed to start: %v", err)
	}
}
