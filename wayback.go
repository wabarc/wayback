// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package wayback // import "github.com/wabarc/wayback"

import (
	"github.com/wabarc/archive.is/pkg"
	"github.com/wabarc/archive.org/pkg"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/logger"
	"github.com/wabarc/wbipfs"
)

type Archived map[string]string

type Broker interface {
	IA() Archived
	IS() Archived
	IP() Archived
}

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
		logger.Error("Wayback %v to Archive.today failed, %v", h.URLs, err)
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
