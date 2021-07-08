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
	aurora "github.com/logrusorgru/aurora/v3"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/storage"
)

type Tor struct {
	ctx   context.Context
	pool  pooling.Pool
	store *storage.Storage
}

// New tor struct.
func New(ctx context.Context, store *storage.Storage, pool pooling.Pool) *Tor {
	if store == nil {
		logger.Fatal("[web] must initialize storage")
	}
	if pool == nil {
		logger.Fatal("[web] must initialize pooling")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	return &Tor{
		ctx:   ctx,
		pool:  pool,
		store: store,
	}
}

// Serve accepts incoming HTTP requests over Tor network, or open
// a local port for proxy server by "WAYBACK_TOR_LOCAL_PORT" env.
// Use "WAYBACK_TOR_PRIVKEY" to keep the Tor hidden service hostname.
//
// Serve always returns an error.
func (t *Tor) Serve() error {
	// Start tor with some defaults + elevated verbosity
	logger.Info("[web] starting and registering onion service, please wait a bit...")

	if _, err := exec.LookPath("tor"); err != nil {
		logger.Fatal("%v", err)
	}

	var pvk ed25519.PrivateKey
	if config.Opts.TorPrivKey() == "" {
		if keypair, err := ed25519.GenerateKey(rand.Reader); err != nil {
			logger.Fatal("[web] generate key failed: %v", err)
		} else {
			pvk = keypair.PrivateKey()
		}
		logger.Info("[web] important to keep the private key: %s", aurora.Blue(hex.EncodeToString(pvk)))
	} else {
		privb, err := hex.DecodeString(config.Opts.TorPrivKey())
		if err != nil {
			logger.Fatal("[web] the key %s is not specific", err)
		}
		pvk = ed25519.PrivateKey(privb)
	}

	verbose := config.Opts.HasDebugMode()
	// startConf := &tor.StartConf{ProcessCreator: libtor.Creator, DataDir: "tor-data"}
	startConf := &tor.StartConf{TorrcFile: t.torrc(), TempDataDirBase: os.TempDir()}
	if verbose {
		startConf.DebugWriter = os.Stdout
	} else {
		startConf.ExtraArgs = []string{"--quiet"}
	}
	e, err := tor.Start(t.ctx, startConf)
	if err != nil {
		logger.Fatal("[web] failed to start tor: %v", err)
	}
	defer e.Close()
	e.DeleteDataDirOnClose = true
	e.StopProcessOnClose = false

	// Create an onion service to listen on any port but show as local port,
	// specify the local port using the `WAYBACK_TOR_LOCAL_PORT` environment variable.
	onion, err := e.Listen(t.ctx, &tor.ListenConf{LocalPort: config.Opts.TorLocalPort(), RemotePorts: config.Opts.TorRemotePorts(), Version3: true, Key: pvk})
	if err != nil {
		logger.Fatal("[web] failed to create onion service: %v", err)
	}
	defer onion.Close()
	onion.CloseLocalListenerOnClose = false

	logger.Info(`[web] listening on "%s" without TLS`, aurora.Blue(onion.LocalListener.Addr()))
	logger.Info("[web] please open a Tor capable browser and navigate to http://%v.onion", onion.ID)

	server := http.Server{Handler: newWeb().handle(t.pool)}
	go func() {
		if err := server.Serve(onion); err != nil {
			logger.Error("[web] serve tor hidden service failed: %v", err)
		}
	}()

	<-t.ctx.Done()
	logger.Info("[web] stopping tor hidden service...")
	if err := server.Shutdown(t.ctx); err != nil {
		logger.Error("[web] shutdown tor hidden service failed: %v", err)
		return err
	}

	return errors.New("done")
}

func (t *Tor) torrc() string {
	if config.Opts.TorrcFile() == "" {
		return ""
	}
	if torPortBusy() {
		return ""
	}
	if _, err := os.Open(config.Opts.TorrcFile()); err != nil {
		return ""
	}
	return config.Opts.TorrcFile()
}

func torPortBusy() bool {
	addr := net.JoinHostPort("127.0.0.1", "9050")
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		logger.Warn("[web] defaults tor port is idle")
		return false
	}
	if conn != nil {
		conn.Close()
		logger.Warn("[web] defaults tor port is busy")
		return true
	}

	return false
}
