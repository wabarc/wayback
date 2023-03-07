// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package httpd // import "github.com/wabarc/wayback/service/httpd"

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	// "github.com/ipsn/go-libtor"
	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
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

// Tor represents a Tor service in the application.
type Tor struct {
	ctx   context.Context
	pub   *publish.Publish
	opts  *config.Options
	pool  *pooling.Pool
	store *storage.Storage

	tor    *tor.Tor
	server *http.Server
}

// New tor struct.
func New(ctx context.Context, opts service.Options) *Tor {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Tor{
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
func (t *Tor) Serve() error {
	// Start tor with some defaults + elevated verbosity
	logger.Info("starting and registering onion service, please wait a bit...")

	handler := newWeb(t.ctx, t.opts, t.pool, t.pub).handle()
	server := &http.Server{
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  5 * time.Minute,
		Handler:      handler,
	}

	switch {
	case torExist():
		logger.Info("start a tor hidden server")
		t.startTorServer(server)
	default:
		logger.Info("start a clear web server")
		server.Addr = t.opts.ListenAddr()
		go startHTTPServer(server)
		t.server = server
	}

	// Block until context done
	<-t.ctx.Done()

	return ErrServiceClosed
}

// Shutdown shuts down the Tor server
func (t *Tor) Shutdown() error {
	// Close onion service.
	if t.tor != nil {
		if err := t.tor.Close(); err != nil {
			return err
		}
	}
	// Shutdown http server
	if t.server != nil {
		if err := t.server.Shutdown(t.ctx); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tor) startTorServer(server *http.Server) {
	var pvk ed25519.PrivateKey
	if t.opts.TorPrivKey() == "" {
		if keypair, err := ed25519.GenerateKey(rand.Reader); err != nil {
			logger.Fatal("generate key failed: %v", err)
		} else {
			pvk = keypair.PrivateKey()
		}
		logger.Info("important to keep the private key: %s", color.BlueString(hex.EncodeToString(pvk)))
	} else {
		privb, err := hex.DecodeString(t.opts.TorPrivKey())
		if err != nil {
			logger.Fatal("the key %s is not specific", err)
		}
		pvk = ed25519.PrivateKey(privb)
	}

	verbose := t.opts.HasDebugMode()
	// startConf := &tor.StartConf{ProcessCreator: libtor.Creator, DataDir: "tor-data"}
	startConf := &tor.StartConf{TempDataDirBase: os.TempDir()}
	if verbose {
		startConf.DebugWriter = os.Stdout
	} else {
		startConf.ExtraArgs = []string{"--quiet"}
	}
	e, err := tor.Start(t.ctx, startConf)
	if err != nil {
		logger.Fatal("failed to start tor: %v", err)
	}
	e.DeleteDataDirOnClose = true
	e.StopProcessOnClose = false

	// Assign e to Tor.tor
	t.tor = e

	listener, err := net.Listen("tcp", t.opts.ListenAddr())
	if err != nil {
		logger.Warn("failed to create local network listener: %v", err)
	}

	// Create an onion service to listen on any port but show as local port,
	// specify the local port using the `WAYBACK_TOR_LOCAL_PORT` environment variable.
	onion, err := e.Listen(t.ctx, &tor.ListenConf{
		LocalPort:     t.opts.TorLocalPort(),
		LocalListener: listener,
		RemotePorts:   t.opts.TorRemotePorts(),
		Version3:      true,
		Key:           pvk,
	})
	if err != nil {
		logger.Fatal("failed to create onion service: %v", err)
	}
	onion.CloseLocalListenerOnClose = true

	logger.Info(`listening on "%s" without TLS`, color.BlueString(onion.LocalListener.Addr().String()))
	logger.Info("please open a Tor capable browser and navigate to http://%v.onion", onion.ID)

	go func() {
		if err := server.Serve(onion); err != nil {
			logger.Fatal("serve tor hidden service failed: %v", err)
		}
	}()
}

func startHTTPServer(server *http.Server) {
	logger.Info(`Listening on "%s" without TLS`, color.BlueString(server.Addr))
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal("Server failed to start: %v", err)
	}
}

func torExist() bool {
	if _, err := exec.LookPath("tor"); err != nil {
		return false
	}
	return true
}
