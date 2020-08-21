package wayback

import (
	"github.com/wabarc/archive.is/pkg"
	"github.com/wabarc/archive.org/pkg"
	"github.com/wabarc/wbipfs"
)

type goal map[string]string

type Broker interface {
	IA() goal
	IS() goal
	WBIPFS() goal
}

type Handle struct {
	URI  []string
	IPFS *IPFSRV
	goal map[string]string
}

func (h *Handle) IA() goal {
	wbrc := &ia.Archiver{}
	uris, _ := wbrc.Wayback(h.URI)

	return uris
}

func (h *Handle) IS() goal {
	wbrc := &is.Archiver{}
	uris, _ := wbrc.Wayback(h.URI)

	return uris
}

func (h *Handle) WBIPFS() goal {
	wbrc := &wbipfs.Archiver{IPFSHost: h.IPFS.Host, IPFSPort: h.IPFS.Port, IPFSMode: "pinner", UseTor: h.IPFS.UseTor}
	uris, _ := wbrc.Wayback(h.URI)

	return uris
}
