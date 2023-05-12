// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package wayback // import "github.com/wabarc/wayback"

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/wabarc/logger"
	"github.com/wabarc/playback"
	"github.com/wabarc/proxier"
	"github.com/wabarc/rivet/ipfs"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/reduxer"
	"golang.org/x/sync/errgroup"

	is "github.com/wabarc/archive.is"
	ia "github.com/wabarc/archive.org"
	ip "github.com/wabarc/rivet"
	ph "github.com/wabarc/telegra.ph"

	pinner "github.com/wabarc/ipfs-pinner"
)

// TODO: find a better way to handle it
var client = &http.Client{}

// Collect results that archived, Arc is name of the archive service,
// Dst mapping the original URL and archived destination URL,
// Ext is extra descriptions.
type Collect struct {
	Arc string // Archive slot name, see config/config.go
	Dst string // Archived destination URL
	Src string // Source URL
	Ext string // Extra identifier
}

// IA represents the Internet Archive slot.
type IA struct {
	ctx context.Context
	cfg *config.Options
	URL *url.URL
}

// IS represents the archive.today slot.
type IS struct {
	ctx context.Context
	cfg *config.Options
	URL *url.URL
}

// IP represents the IPFS slot.
type IP struct {
	ctx context.Context
	cfg *config.Options
	URL *url.URL
}

// PH represents the Telegra.ph slot.
type PH struct {
	ctx context.Context
	cfg *config.Options
	URL *url.URL
}

// Waybacker is the interface that wraps the basic Wayback method.
//
// Wayback wayback *url.URL from struct of the implementations to the Wayback Machine.
// It returns the result of string from the upstream services.
type Waybacker interface {
	Wayback(reduxer.Reduxer) string
}

// Wayback implements the standard Waybacker interface:
// it reads URL from the IA and returns archived URL as a string.
func (i IA) Wayback(_ reduxer.Reduxer) string {
	arc := &ia.Archiver{Client: client}
	dst, err := arc.Wayback(i.ctx, i.URL)
	if err != nil {
		logger.Error("wayback %s to Internet Archive failed: %v", i.URL.String(), err)
		return fmt.Sprint(err)
	}
	return dst
}

// Wayback implements the standard Waybacker interface:
// it reads URL from the IS and returns archived URL as a string.
func (i IS) Wayback(_ reduxer.Reduxer) string {
	arc := is.NewArchiver(client)
	defer arc.CloseTor()

	dst, err := arc.Wayback(i.ctx, i.URL)
	if err != nil {
		logger.Error("wayback %s to archive.today failed: %v", i.URL.String(), err)
		return fmt.Sprint(err)
	}
	return dst
}

// Wayback implements the standard Waybacker interface:
// it reads URL from the IP and returns archived URL as a string.
func (i IP) Wayback(rdx reduxer.Reduxer) string {
	opts := []ipfs.PinningOption{
		ipfs.Mode(ipfs.Remote),
	}
	if i.cfg.IPFSMode() == "daemon" {
		opts = []ipfs.PinningOption{
			ipfs.Mode(ipfs.Local),
			ipfs.Host(i.cfg.IPFSHost()),
			ipfs.Port(i.cfg.IPFSPort()),
		}
	}

	target := i.cfg.IPFSTarget()
	switch target {
	case pinner.Infura, pinner.Pinata, pinner.NFTStorage, pinner.Web3Storage:
		apikey := i.cfg.IPFSApikey()
		secret := i.cfg.IPFSSecret()
		opts = append(opts, ipfs.Uses(target), ipfs.Apikey(apikey), ipfs.Secret(secret))
	}
	arc := &ip.Shaft{Hold: ipfs.Options(opts...)}
	uri := i.URL.String()
	ctx := i.ctx

	// If there is bundled HTML, it is utilized as the basis for IPFS
	// archiving and is sent to obelisk to crawl the rest of the page.
	if bundle, ok := rdx.Load(reduxer.Src(uri)); ok {
		shot := bundle.Shots()
		buf, err := os.ReadFile(fmt.Sprint(shot.HTML))
		if err == nil {
			ctx = arc.WithInput(ctx, buf)
		}
	}

	dst, err := arc.Wayback(ctx, i.URL)
	if err != nil {
		logger.Error("wayback %s to IPFS failed: %v", i.URL.String(), err)
		return fmt.Sprint(err)
	}
	return dst
}

// Wayback implements the standard Waybacker interface:
// it reads URL from the PH and returns archived URL as a string.
func (i PH) Wayback(rdx reduxer.Reduxer) string {
	arc := ph.New(client)
	uri := i.URL.String()
	ctx := i.ctx

	if i.cfg.EnabledChromeRemote() {
		arc.ByRemote(i.cfg.ChromeRemoteAddr())
	}
	if bundle, ok := rdx.Load(reduxer.Src(uri)); ok {
		ctx = arc.WithShot(ctx, bundle.Shots())
		ctx = arc.WithArticle(ctx, bundle.Article())
	}

	dst, err := arc.Wayback(ctx, i.URL)
	if err != nil {
		logger.Error("wayback %s to telegra.ph failed: %v", i.URL.String(), err)
		return fmt.Sprint(err)
	}
	return dst
}

func wayback(w Waybacker, r reduxer.Reduxer) string {
	return w.Wayback(r)
}

// Wayback returns URLs archived to the time capsules of given URLs.
func Wayback(ctx context.Context, rdx reduxer.Reduxer, cfg *config.Options, urls ...*url.URL) ([]Collect, error) {
	logger.Debug("start...")

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.WaybackTimeout())
		defer cancel()
	}
	ctx, cancel := context.WithTimeout(ctx, duration(ctx))
	defer cancel()

	mu := sync.Mutex{}
	cols := []Collect{}
	g, ctx := errgroup.WithContext(ctx)
	for _, input := range urls {
		for slot, arc := range cfg.Slots() {
			if !arc {
				logger.Warn("skipped %s", config.SlotName(slot))
				continue
			}
			slot, input := slot, input
			g.Go(func() error {
				logger.Debug("archiving slot: %s", slot)

				uri := input.String()
				var col Collect
				switch slot {
				case config.SLOT_IA:
					col.Dst = wayback(IA{URL: input, cfg: cfg, ctx: ctx}, rdx)
				case config.SLOT_IS:
					col.Dst = wayback(IS{URL: input, cfg: cfg, ctx: ctx}, rdx)
				case config.SLOT_IP:
					col.Dst = wayback(IP{URL: input, cfg: cfg, ctx: ctx}, rdx)
				case config.SLOT_PH:
					col.Dst = wayback(PH{URL: input, cfg: cfg, ctx: ctx}, rdx)
				}
				col.Src = uri
				col.Arc = slot
				col.Ext = slot
				mu.Lock()
				cols = append(cols, col)
				mu.Unlock()
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		logger.Error("archiving some slot unexpected: %v", err)
	}

	if len(cols) == 0 {
		return cols, errors.New("archiving failed: no cols")
	}

	return cols, nil
}

// Playback returns URLs archived from the time capsules.
func Playback(ctx context.Context, cfg *config.Options, urls ...*url.URL) (cols []Collect, err error) {
	logger.Debug("start...")

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.WaybackTimeout())
		defer cancel()
	}
	ctx, cancel := context.WithTimeout(ctx, duration(ctx))
	defer cancel()

	mu := sync.Mutex{}
	g, ctx := errgroup.WithContext(ctx)
	var slots = []string{config.SLOT_IA, config.SLOT_IS, config.SLOT_IP, config.SLOT_PH, config.SLOT_TT, config.SLOT_GC}
	for _, input := range urls {
		for _, slot := range slots {
			slot, input := slot, input
			g.Go(func() error {
				logger.Debug("searching slot: %s", slot)
				var col Collect
				switch slot {
				case config.SLOT_IA:
					col.Dst = playback.Playback(ctx, playback.IA{URL: input})
				case config.SLOT_IS:
					col.Dst = playback.Playback(ctx, playback.IS{URL: input})
				case config.SLOT_IP:
					col.Dst = playback.Playback(ctx, playback.IP{URL: input})
				case config.SLOT_PH:
					col.Dst = playback.Playback(ctx, playback.PH{URL: input})
				case config.SLOT_TT:
					col.Dst = playback.Playback(ctx, playback.TT{URL: input})
				case config.SLOT_GC:
					col.Dst = playback.Playback(ctx, playback.GC{URL: input})
				}
				col.Src = input.String()
				col.Arc = slot
				col.Ext = slot
				mu.Lock()
				cols = append(cols, col)
				mu.Unlock()
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		logger.Error("playback some slot unexpected: %v", err)
	}

	if len(cols) == 0 {
		return cols, errors.New("playback failed: no cols")
	}

	return cols, nil
}

// duration reduce the context deadline time for downstream and reserve
// extra time for the caller.
func duration(ctx context.Context) time.Duration {
	deadline, _ := ctx.Deadline()
	elapsed := deadline.Unix() - time.Now().Unix()
	safeTime := elapsed * 90 / 100

	return time.Duration(safeTime) * time.Second
}

// NewClient sets a http client.
// TODO: refactoring
func SetClient(ctx context.Context, opts *config.Options) {
	if opts.Proxy() != "" {
		var (
			err    error
			cancel context.CancelFunc
		)
		client.Transport, err = proxier.NewUTLSRoundTripper(proxier.Proxy(opts.Proxy()))
		if err != nil {
			logger.Error("create utls round tripper failed: %v", err)
			return
		}
		if opts.HasDebugMode() {
			endpoint := "https://icanhazip.com"
			go func() {
				for {
					ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
					req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil) // nolint:errcheck
					resp, err := client.Do(req)
					if err != nil {
						logger.Error("request error: %v", err)
						cancel()
						continue
					}
					cancel()
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						logger.Error("read body error: %v", err)
						continue
					}
					logger.Debug("client handshake: %s", bytes.TrimSpace(body))
					resp.Body.Close()
					time.Sleep(time.Minute)
				}
			}()
		}
	}
}
