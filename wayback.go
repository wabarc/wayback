// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package wayback // import "github.com/wabarc/wayback"

import (
	"github.com/wabarc/archive.is/pkg"
	"github.com/wabarc/archive.org/pkg"
	"github.com/wabarc/telegra.ph/pkg"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/logger"
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

// Handle URLs need to wayback and configs,
// Opts on `github.com/wabarc/wayback/config`.
type Handle struct {
	URLs []string

	Opts *config.Options
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
		IPFSHost: h.Opts.IPFSHost(),
		IPFSPort: h.Opts.IPFSPort(),
		IPFSMode: h.Opts.IPFSMode(),
		UseTor:   h.Opts.UseTor(),
	}
	uris, err := wbrc.Wayback(h.URLs)
	if err != nil {
		logger.Error("Wayback %v to IPFS failed, %v", h.URLs, err)
	}

	return uris
}

func (h *Handle) PH() Archived {
	wbrc := &ph.Archiver{}
	uris, err := wbrc.Wayback(h.URLs)
	if err != nil {
		logger.Error("Wayback %v to Telegra.ph failed, %v", h.URLs, err)
	}

	return uris
}
