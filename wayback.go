// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package wayback // import "github.com/wabarc/wayback"

import (
	"context"
	"sync"

	is "github.com/wabarc/archive.is"
	ia "github.com/wabarc/archive.org"
	"github.com/wabarc/logger"
	"github.com/wabarc/playback"
	"github.com/wabarc/screenshot"
	ph "github.com/wabarc/telegra.ph"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wbipfs"
)

// Archived returns result of wayback.
type Archived map[string]string

// Broker is interface of the wayback,
// methods returns `Archived`.
type Broker interface {
	IA() Archived
	IS() Archived
	IP() Archived
	PH() Archived
}

// Handle represents a wayback handle.
type Handle struct {
	Bundles *[]reduxer.Bundle

	URLs []string
}

// Collect result that archived, Arc is name of the archive service,
// Dst mapping the original URL and archived destination URL,
// Ext is extra descriptions.
type Collect struct {
	Arc string            // Archive slot name, see config/config.go
	Dst map[string]string // Archived destination URL
	Ext string            // Extra identifier
}

func (h *Handle) IA() Archived {
	wbrc := &ia.Archiver{}
	uris, err := wbrc.Wayback(h.URLs)
	if err != nil {
		logger.Error("Wayback %v to Internet Archive failed, %v", h.URLs, err)
	}

	return uris
}

func (h *Handle) IS() Archived {
	wbrc := &is.Archiver{}
	uris, err := wbrc.Wayback(h.URLs)
	if err != nil {
		logger.Error("Wayback %v to archive.today failed, %v", h.URLs, err)
	}

	return uris
}

func (h *Handle) IP() Archived {
	wbrc := &wbipfs.Archiver{
		IPFSHost: config.Opts.IPFSHost(),
		IPFSPort: config.Opts.IPFSPort(),
		IPFSMode: config.Opts.IPFSMode(),
		UseTor:   config.Opts.UseTor(),
	}
	uris, err := wbrc.Wayback(h.URLs)
	if err != nil {
		logger.Error("Wayback %v to IPFS failed, %v", h.URLs, err)
	}

	return uris
}

func (h *Handle) PH() Archived {
	wbrc := &ph.Archiver{}
	wbrc.SetShots(h.parseShots())
	if config.Opts.EnabledChromeRemote() {
		wbrc.ByRemote(config.Opts.ChromeRemoteAddr())
	}

	uris, err := wbrc.Wayback(h.URLs)
	if err != nil {
		logger.Error("Wayback %v to Telegra.ph failed, %v", h.URLs, err)
	}

	return uris
}

func (h *Handle) parseShots() (s []screenshot.Screenshots) {
	for _, bundle := range *h.Bundles {
		s = append(s, screenshot.Screenshots{
			URL:   bundle.URL,
			Title: bundle.Title,
			Image: bundle.Image,
			HTML:  bundle.HTML,
			PDF:   bundle.PDF,
		})
	}
	return s
}

// Wayback returns URLs archived to the time capsules.
func Wayback(urls []string, bundles *[]reduxer.Bundle) (col []Collect, err error) {
	logger.Debug("[wayback] start...")

	*bundles, err = reduxer.Do(context.Background(), urls...)
	if err != nil {
		logger.Info("[wayback] cannot to start reduxer: %v", err)
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	var wb Broker = &Handle{URLs: urls, Bundles: bundles}
	for slot, arc := range config.Opts.Slots() {
		if !arc {
			continue
		}
		wg.Add(1)
		go func(slot string) {
			defer wg.Done()
			c := Collect{}
			logger.Debug("[wayback] archiving slot: %s", slot)
			switch slot {
			case config.SLOT_IA:
				c.Dst = wb.IA()
			case config.SLOT_IS:
				c.Dst = wb.IS()
			case config.SLOT_IP:
				c.Dst = wb.IP()
			case config.SLOT_PH:
				c.Dst = wb.PH()
			}
			c.Arc = config.SlotName(slot)
			c.Ext = config.SlotExtra(slot)
			mu.Lock()
			col = append(col, c)
			mu.Unlock()
		}(slot)
	}
	wg.Wait()

	if len(col) == 0 {
		logger.Error("[wayback] archives failure")
		return col, errors.New("archives failure")
	}

	return col, nil
}

// Playback returns URLs archived from the time capsules.
func Playback(urls []string) (col []Collect, err error) {
	logger.Debug("[playback] start...")

	var mu sync.Mutex
	var wg sync.WaitGroup
	var pb playback.Playback = &playback.Handle{URLs: urls}
	var slots = []string{config.SLOT_IA, config.SLOT_IS, config.SLOT_IP, config.SLOT_PH, config.SLOT_TT, config.SLOT_GC}
	for _, slot := range slots {
		wg.Add(1)
		go func(slot string) {
			defer wg.Done()
			c := Collect{}
			logger.Debug("[playback] searching slot: %s", slot)
			switch slot {
			case config.SLOT_IA:
				c.Dst = pb.IA()
			case config.SLOT_IS:
				c.Dst = pb.IS()
			case config.SLOT_IP:
				c.Dst = pb.IP()
			case config.SLOT_PH:
				c.Dst = pb.PH()
			case config.SLOT_TT:
				c.Dst = pb.TT()
			case config.SLOT_GC:
				c.Dst = pb.GC()
			}
			c.Arc = config.SlotName(slot)
			c.Ext = config.SlotExtra(slot)
			mu.Lock()
			col = append(col, c)
			mu.Unlock()
		}(slot)
	}
	wg.Wait()

	return col, nil
}
