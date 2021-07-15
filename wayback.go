// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package wayback // import "github.com/wabarc/wayback"

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/wabarc/logger"
	"github.com/wabarc/playback"
	"github.com/wabarc/screenshot"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/reduxer"
	"golang.org/x/sync/errgroup"

	is "github.com/wabarc/archive.is"
	ia "github.com/wabarc/archive.org"
	ph "github.com/wabarc/telegra.ph"
	ip "github.com/wabarc/wbipfs"
)

// Collect results that archived, Arc is name of the archive service,
// Dst mapping the original URL and archived destination URL,
// Ext is extra descriptions.
type Collect struct {
	Arc string // Archive slot name, see config/config.go
	Dst string // Archived destination URL
	Src string // Source URL
	Ext string // Extra identifier
}

type IA struct {
	URL *url.URL
	ctx context.Context
}

type IS struct {
	ctx context.Context

	URL *url.URL
}

type IP struct {
	ctx context.Context

	URL *url.URL
}

type PH struct {
	ctx    context.Context
	bundle reduxer.Bundle

	URL *url.URL
}

// Waybacker is the interface that wraps the basic Wayback method.
//
// Wayback wayback *url.URL from struct of the implementations to the Wayback Machine.
// It returns the result of string from the upstream services.
type Waybacker interface {
	Wayback() string
}

func (i IA) Wayback() string {
	arc := &ia.Archiver{}
	dst, err := arc.Wayback(i.ctx, i.URL)
	if err != nil {
		logger.Error("[wayback] %s to Internet Archive failed: %v", i.URL.String(), err)
		return fmt.Sprint(err)
	}
	return dst
}

func (i IS) Wayback() string {
	arc := &is.Archiver{}
	dst, err := arc.Wayback(i.ctx, i.URL)
	if err != nil {
		logger.Error("[wayback] %s to archive.today failed: %v", i.URL.String(), err)
		return fmt.Sprint(err)
	}
	return dst
}

func (i IP) Wayback() string {
	arc := &ip.Archiver{
		IPFSHost: config.Opts.IPFSHost(),
		IPFSPort: config.Opts.IPFSPort(),
		IPFSMode: config.Opts.IPFSMode(),
		UseTor:   config.Opts.UseTor(),
	}
	dst, err := arc.Wayback(i.ctx, i.URL)
	if err != nil {
		logger.Error("[wayback] %s to IPFS failed: %v", i.URL.String(), err)
		return fmt.Sprint(err)
	}
	return dst
}

func (i PH) Wayback() string {
	arc := &ph.Archiver{}
	arc.SetShot(i.parseShot())
	if config.Opts.EnabledChromeRemote() {
		arc.ByRemote(config.Opts.ChromeRemoteAddr())
	}

	dst, err := arc.Wayback(i.ctx, i.URL)
	if err != nil {
		logger.Error("[wayback] %s to archive.today failed: %v", i.URL.String(), err)
		return fmt.Sprint(err)
	}
	return dst
}

func (p PH) parseShot() (shot screenshot.Screenshots) {
	return screenshot.Screenshots{
		URL:   p.bundle.URL,
		Title: p.bundle.Title,
		Image: p.bundle.Image,
		HTML:  p.bundle.HTML,
		PDF:   p.bundle.PDF,
	}
}

func wayback(w Waybacker) string {
	return w.Wayback()
}

// Wayback returns URLs archived to the time capsules.
func Wayback(ctx context.Context, bundles *reduxer.Bundles, urls ...string) (cols []Collect, err error) {
	logger.Debug("[wayback] start...")

	*bundles, err = reduxer.Do(ctx, urls...)
	if err != nil {
		logger.Warn("[wayback] cannot to start reduxer: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	mu := sync.Mutex{}
	g, ctx := errgroup.WithContext(ctx)
	for _, uri := range urls {
		for slot, arc := range config.Opts.Slots() {
			if !arc {
				logger.Warn("[wayback] skipped %s", config.SlotName(slot))
				continue
			}
			slot, uri := slot, uri
			g.Go(func() error {
				logger.Debug("[wayback] archiving slot: %s", slot)
				input, err := url.Parse(uri)
				if err != nil {
					logger.Error("[wayback] parse uri failed: %v", err)
					return err
				}

				bundle := (*bundles)[uri]
				var col Collect
				switch slot {
				case config.SLOT_IA:
					col.Dst = wayback(IA{URL: input, ctx: ctx})
				case config.SLOT_IS:
					col.Dst = wayback(IS{URL: input, ctx: ctx})
				case config.SLOT_IP:
					col.Dst = wayback(IP{URL: input, ctx: ctx})
				case config.SLOT_PH:
					col.Dst = wayback(PH{URL: input, ctx: ctx, bundle: bundle})
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
		logger.Error("[wayback] archives failed: %v", err)
		return cols, err
	}

	if len(cols) == 0 {
		logger.Error("[wayback] archives failure")
		return cols, errors.New("archives failure")
	}

	return cols, nil
}

// Playback returns URLs archived from the time capsules.
func Playback(ctx context.Context, urls ...string) (cols []Collect, err error) {
	logger.Debug("[playback] start...")

	mu := sync.Mutex{}
	g, ctx := errgroup.WithContext(ctx)
	var slots = []string{config.SLOT_IA, config.SLOT_IS, config.SLOT_IP, config.SLOT_PH, config.SLOT_TT, config.SLOT_GC}
	for _, uri := range urls {
		for _, slot := range slots {
			slot, uri := slot, uri
			g.Go(func() error {
				logger.Debug("[playback] searching slot: %s", slot)
				input, err := url.Parse(uri)
				if err != nil {
					logger.Error("[playback] parse uri failed: %v", err)
					return err
				}
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
		logger.Error("[playback] failed: %v", err)
		return cols, err
	}

	return cols, nil
}
