// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package anonymity // import "github.com/wabarc/wayback/service/anonymity"

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"os/exec"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
	// "github.com/ipsn/go-libtor"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
)

type Tor struct {
}

// New tor struct.
func New() *Tor {
	return &Tor{}
}

// Serve accepts incoming HTTP requests over Tor network, or open
// a local port for proxy server by "WAYBACK_TOR_LOCAL_PORT" env.
// Use "WAYBACK_TOR_PRIVKEY" to keep the Tor hidden service hostname.
//
// Serve always returns an error.
func (t *Tor) Serve(ctx context.Context) error {
	// Start tor with some defaults + elevated verbosity
	logger.Info("[web] starting and registering onion service, please wait a bit...")

	if _, err := exec.LookPath("tor"); err != nil {
		logger.Fatal("%v", err)
	}

	var pvk ed25519.PrivateKey
	if config.Opts.TorPrivKey() == "" {
		keypair, _ := ed25519.GenerateKey(rand.Reader)
		pvk = keypair.PrivateKey()
		logger.Info("[web] important to keep the private key: %s", hex.EncodeToString(pvk))
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
	e, err := tor.Start(ctx, startConf)
	if err != nil {
		logger.Fatal("[web] failed to start tor: %v", err)
	}
	defer e.Close()
	e.DeleteDataDirOnClose = true
	e.StopProcessOnClose = false

	// Create an onion service to listen on any port but show as local port,
	// specify the local port using the `WAYBACK_TOR_LOCAL_PORT` environment variable.
	onion, err := e.Listen(ctx, &tor.ListenConf{LocalPort: config.Opts.TorLocalPort(), RemotePorts: config.Opts.TorRemotePorts(), Version3: true, Key: pvk})
	if err != nil {
		logger.Fatal("[web] failed to create onion service: %v", err)
	}
	defer onion.Close()
	onion.CloseLocalListenerOnClose = false

	logger.Info(`[web] listening on %q without TLS`, onion.LocalListener.Addr())
	logger.Info("[web] please open a Tor capable browser and navigate to http://%v.onion", onion.ID)

	server := http.Server{Handler: newWeb().handle()}
	go func() {
		server.Serve(onion)
	}()

	select {
	case <-ctx.Done():
		logger.Info("[web] stopping tor hidden service...")
		server.Shutdown(ctx)
	}

	return errors.New("done")
}

func (t *Tor) torrc() string {
	if config.Opts.TorrcFile() == "" {
		return ""
	}
	if _, err := os.Open(config.Opts.TorrcFile()); err != nil {
		return ""
	}
	return config.Opts.TorrcFile()
}
