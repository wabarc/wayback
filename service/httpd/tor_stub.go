// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

//go:build with_tor

package httpd // import "github.com/wabarc/wayback/service/httpd"

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"os"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
	"github.com/fatih/color"
	"github.com/ipsn/go-libtor"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/errors"
)

var creator = libtor.Creator

func (h *Httpd) startTorServer(server *http.Server) error {
	var pvk ed25519.PrivateKey
	if h.opts.TorPrivKey() == "" {
		if keypair, err := ed25519.GenerateKey(rand.Reader); err != nil {
			return errors.Wrap(err, "generate key failed")
		} else {
			pvk = keypair.PrivateKey()
		}
		logger.Info("important to keep the private key: %s", color.BlueString(hex.EncodeToString(pvk)))
	} else {
		privb, err := hex.DecodeString(h.opts.TorPrivKey())
		if err != nil {
			return errors.Wrap(err, "key is not specific")
		}
		pvk = ed25519.PrivateKey(privb)
	}

	verbose := h.opts.HasDebugMode()
	startConf := &tor.StartConf{ProcessCreator: creator, TempDataDirBase: os.TempDir()}
	if verbose {
		startConf.DebugWriter = os.Stdout
	} else {
		startConf.ExtraArgs = []string{"--quiet"}
	}
	e, err := tor.Start(h.ctx, startConf)
	if err != nil {
		return errors.Wrap(err, "failed to start tor")
	}
	e.DeleteDataDirOnClose = true
	e.StopProcessOnClose = false

	// Assign e to Tor.tor
	h.tor = e

	listener, err := net.Listen("tcp", h.opts.ListenAddr())
	if err != nil {
		logger.Warn("failed to create local network listener: %v", err)
	}

	// Create an onion service to listen on any port but show as local port,
	// specify the local port using the `WAYBACK_TOR_LOCAL_PORT` environment variable.
	onion, err := e.Listen(h.ctx, &tor.ListenConf{
		LocalPort:     h.opts.TorLocalPort(),
		LocalListener: listener,
		RemotePorts:   h.opts.TorRemotePorts(),
		Version3:      true,
		Key:           pvk,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create onion service")
	}
	onion.CloseLocalListenerOnClose = true

	logger.Info(`listening on "%s" without TLS`, color.BlueString(onion.LocalListener.Addr().String()))
	logger.Info("please open a Tor capable browser and navigate to http://%v.onion", onion.ID)

	go func() {
		if err := server.Serve(onion); err != nil {
			logger.Fatal("serve tor hidden service failed: %v", err)
		}
	}()

	return nil
}

func torExist() bool {
	return true
}
