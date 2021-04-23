// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package anonymity // import "github.com/wabarc/wayback/service/anonymity"

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
	// "github.com/ipsn/go-libtor"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/template"
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

	// Create an onion service to listen on any port but show as local port,
	// specify the local port using the `WAYBACK_TOR_LOCAL_PORT` environment variable.
	onion, err := e.Listen(ctx, &tor.ListenConf{LocalPort: config.Opts.TorLocalPort(), RemotePorts: config.Opts.TorRemotePorts(), Version3: true, Key: pvk})
	if err != nil {
		logger.Fatal("[web] failed to create onion service: %v", err)
	}
	defer onion.Close()

	logger.Info(`[web] listening on %q without TLS`, onion.LocalListener.Addr())
	logger.Info("[web] please open a Tor capable browser and navigate to http://%v.onion", onion.ID)

	go func() {
		http.Serve(onion, newWeb().handle())
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	return errors.New("done")
}

func (web *web) process(w http.ResponseWriter, r *http.Request) {
	logger.Debug("[web] process request start...")
	if r.Method != http.MethodPost {
		logger.Info("[web] request method no specific.")
		http.Redirect(w, r, "/", http.StatusNotModified)
		return
	}

	if err := r.ParseForm(); err != nil {
		logger.Error("[web] parse form error, %v", err)
		http.Redirect(w, r, "/", http.StatusNotModified)
		return
	}

	text := r.PostFormValue("text")
	if len(strings.TrimSpace(text)) == 0 {
		logger.Info("[web] post form value empty.")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	logger.Debug("[web] text: %s", text)

	urls := helper.MatchURL(text)
	if len(urls) == 0 {
		logger.Info("[web] url no found.")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	col, _ := wayback.Wayback(urls)
	collector := transform(col)
	ctx := context.Background()
	switch r.PostFormValue("data-type") {
	case "json":
		w.Header().Set("Content-Type", "application/json")

		if data, err := json.Marshal(collector); err != nil {
			logger.Error("[web] encode for response failed, %v", err)
		} else {
			go publish.To(ctx, col, "web")
			w.Write(data)
		}
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if html, ok := web.template.Render("layout", collector); ok {
			go publish.To(ctx, col, "web")
			w.Write(html)
		} else {
			logger.Error("[web] render template for response failed")
		}
	}
}

func transform(col []*wayback.Collect) template.Collector {
	collects := []template.Collect{}
	for _, c := range col {
		for src, dst := range c.Dst {
			collects = append(collects, template.Collect{
				Slot: c.Arc,
				Src:  src,
				Dst:  dst,
			})
		}
	}
	return collects
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
