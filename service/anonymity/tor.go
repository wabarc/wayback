// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package anonymity // import "github.com/wabarc/wayback/service/anonymity"

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	embedTor "github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
	// "github.com/ipsn/go-libtor"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/logger"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/template"
	"github.com/wabarc/wayback/utils"
)

type tor struct {
	opts *config.Options
}

// New tor struct.
func New(opts *config.Options) *tor {
	return &tor{
		opts: opts,
	}
}

// Serve accepts incoming HTTP requests over Tor network, or open
// a local port for proxy server by "WAYBACK_TOR_LOCAL_PORT" env.
// Use "WAYBACK_TOR_PRIVKEY" to keep the Tor hidden service hostname.
//
// Serve always returns a nil error.
func (t *tor) Serve(ctx context.Context) error {
	// Start tor with some defaults + elevated verbosity
	logger.Info("Web: starting and registering onion service, please wait a bit...")

	if _, err := exec.LookPath("tor"); err != nil {
		logger.Fatal("%v", err)
	}

	var pvk ed25519.PrivateKey
	if t.opts.TorPrivKey() == "" {
		keypair, _ := ed25519.GenerateKey(rand.Reader)
		pvk = keypair.PrivateKey()
		logger.Info("Web: important to keep the private key: %s", hex.EncodeToString(pvk))
	} else {
		privb, err := hex.DecodeString(t.opts.TorPrivKey())
		if err != nil {
			logger.Fatal("Web: the key %s is not specific", err)
		}
		pvk = ed25519.PrivateKey(privb)
	}

	verbose := t.opts.HasDebugMode()
	// startConf := &embedTor.StartConf{ProcessCreator: libtor.Creator, DataDir: "tor-data"}
	startConf := &embedTor.StartConf{}
	if verbose {
		startConf.DebugWriter = os.Stdout
	} else {
		startConf.ExtraArgs = []string{"--quiet"}
	}
	e, err := embedTor.Start(ctx, startConf)
	if err != nil {
		logger.Fatal("Web: failed to start tor: %v", err)
	}
	defer e.Close()

	// Create an onion service to listen on any port but show as 80
	onion, err := e.Listen(ctx, &embedTor.ListenConf{LocalPort: t.opts.TorLocalPort(), RemotePorts: t.opts.TorRemotePorts(), Version3: true, Key: pvk})
	if err != nil {
		logger.Fatal("Web: failed to create onion service: %v", err)
	}
	defer onion.Close()

	logger.Info("Web: please open a Tor capable browser and navigate to http://%v.onion", onion.ID)

	http.HandleFunc("/", home)
	http.HandleFunc("/w", func(w http.ResponseWriter, r *http.Request) { t.process(w, r, ctx) })
	http.Serve(onion, nil)

	return nil
}

func home(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Collector{}
	if html, ok := tmpl.Render(); ok {
		w.Write(html)
	} else {
		logger.Error("Web: render template for home request failed")
		http.Error(w, "Internal Server Error", 500)
	}
}

func (t *tor) process(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	logger.Debug("Web: process request start...")
	if r.Method != http.MethodPost {
		logger.Info("Web: request method no specific.")
		http.Redirect(w, r, "/", 405)
		return
	}

	if err := r.ParseForm(); err != nil {
		logger.Error("Web: parse form error, %v", err)
		http.Redirect(w, r, "/", 400)
		return
	}

	text := r.PostFormValue("text")
	if len(strings.TrimSpace(text)) == 0 {
		logger.Info("Web: post form value empty.")
		http.Redirect(w, r, "/", 411)
		return
	}

	logger.Debug("Web: text: %s", text)

	collector, col := t.archive(ctx, text)
	switch r.PostFormValue("data-type") {
	case "json":
		w.Header().Set("Content-Type", "application/json")

		if data, err := json.Marshal(collector); err != nil {
			logger.Error("Web: encode for response failed, %v", err)
		} else {
			go publish.ToChannel(t.opts, nil, publish.Render(col))
			w.Write(data)
		}

		return
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if html, ok := collector.Render(); ok {
			go publish.ToChannel(t.opts, nil, publish.Render(col))
			w.Write(html)
		} else {
			logger.Error("Web: render template for response failed")
		}

		return
	}
}

func (t *tor) archive(ctx context.Context, text string) (tc *template.Collector, col []*publish.Collect) {
	logger.Debug("Web: archives start...")
	tc = &template.Collector{}

	urls := utils.MatchURL(text)
	if len(urls) == 0 {
		transform(tc, "", map[string]string{text: "URL no found"})
		logger.Info("Web: archives failure, URL no found.")
		return tc, []*publish.Collect{}
	}

	wg := sync.WaitGroup{}
	var wbrc wayback.Broker = &wayback.Handle{URLs: urls, Opts: t.opts}
	for slot, arc := range t.opts.Slots() {
		if !arc {
			continue
		}
		wg.Add(1)
		go func(slot string, tc *template.Collector) {
			defer wg.Done()
			c := &publish.Collect{}
			switch slot {
			case config.SLOT_IA:
				logger.Debug("Web: archiving slot: %s", slot)
				ia := wbrc.IA()
				slotName := config.SlotName(slot)

				// Data for response
				transform(tc, slotName, ia)

				// Data for publish
				c.Arc = fmt.Sprintf("<a href='https://web.archive.org/'>%s</a>", slotName)
				c.Dst = ia
			case config.SLOT_IS:
				logger.Debug("Web: archiving slot: %s", slot)
				is := wbrc.IS()
				slotName := config.SlotName(slot)

				// Data for response
				transform(tc, slotName, is)

				// Data for publish
				c.Arc = fmt.Sprintf("<a href='https://archive.today/'>%s</a>", slotName)
				c.Dst = is
			case config.SLOT_IP:
				logger.Debug("Web: archiving slot: %s", slot)
				ip := wbrc.IP()
				slotName := config.SlotName(slot)

				// Data for response
				transform(tc, slotName, ip)

				// Data for publish
				c.Arc = fmt.Sprintf("<a href='https://ipfs.github.io/public-gateway-checker/'>%s</a>", slotName)
				c.Dst = ip
			}
			col = append(col, c)
		}(slot, tc)
	}
	wg.Wait()

	return tc, col
}

func transform(c *template.Collector, slot string, arc map[string]string) {
	p := *c
	for src, dst := range arc {
		p = append(p, template.Collect{Slot: slot, Src: src, Dst: dst})
	}
	*c = p
}
