// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/reduxer"
)

var collects = []wayback.Collect{
	{
		Arc: config.SLOT_IA,
		Dst: "https://web.archive.org/web/20211000000001/https://example.com/",
		Src: "https://example.com/",
		Ext: config.SLOT_IA,
	},
	{
		Arc: config.SLOT_IS,
		Dst: "http://archive.today/abcdE",
		Src: "https://example.com/",
		Ext: config.SLOT_IS,
	},
	{
		Arc: config.SLOT_IP,
		Dst: "https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr",
		Src: "https://example.com/",
		Ext: config.SLOT_IP,
	},
	{
		Arc: config.SLOT_PH,
		Dst: "http://telegra.ph/title-01-01",
		Src: "https://example.com/",
		Ext: config.SLOT_PH,
	},
}

var flawed = []wayback.Collect{
	{
		Arc: config.SLOT_IA,
		Dst: `Get "https://web.archive.org/save/https://example.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)`,
		Src: "https://example.com/?q=%E4%B8%AD%E6%96%87",
		Ext: config.SLOT_IA,
	},
	{
		Arc: config.SLOT_IS,
		Dst: "http://archive.today/abcdE",
		Src: "https://example.com/",
		Ext: config.SLOT_IS,
	},
	{
		Arc: config.SLOT_IP,
		Dst: "Archive failed.",
		Src: "https://example.com/",
		Ext: config.SLOT_IP,
	},
	{
		Arc: config.SLOT_PH,
		Dst: "https://web.archive.org/*/https://webcache.googleusercontent.com/search?q=cache:https://example.com/404",
		Src: "https://example.com/404",
		Ext: config.SLOT_PH,
	},
}

var multi = []wayback.Collect{
	{
		Arc: config.SLOT_IA,
		Dst: `https://web.archive.org/123/https://example.com/`,
		Src: "https://example.com/",
		Ext: config.SLOT_IA,
	},
	{
		Arc: config.SLOT_IS,
		Dst: "http://archive.today/abcdE",
		Src: "https://example.com/",
		Ext: config.SLOT_IS,
	},
	{
		Arc: config.SLOT_IA,
		Dst: `https://web.archive.org/123/https://example.org/`,
		Src: "https://example.org/",
		Ext: config.SLOT_IA,
	},
	{
		Arc: config.SLOT_IS,
		Dst: "http://archive.today/abc",
		Src: "https://example.org/",
		Ext: config.SLOT_IS,
	},
}

var bundleExample = &reduxer.Bundle{
	Assets: reduxer.Assets{
		Img: reduxer.Asset{
			Local: "/path/to/image",
			Remote: reduxer.Remote{
				Anonfile: "https://anonfiles.com/FbZfSa9eu4",
				Catbox:   "https://files.catbox.moe/9u6yvu.png",
			},
		},
		PDF: reduxer.Asset{
			Local: "/path/to/pdf",
			Remote: reduxer.Remote{
				Anonfile: "https://anonfiles.com/r4G8Sb90ud",
				Catbox:   "https://files.catbox.moe/q73uqh.pdf",
			},
		},
		Raw: reduxer.Asset{
			Local: "/path/to/htm",
			Remote: reduxer.Remote{
				Anonfile: "https://anonfiles.com/pbG4Se94ua",
				Catbox:   "https://files.catbox.moe/bph1g6.htm",
			},
		},
		Txt: reduxer.Asset{
			Local: "/path/to/txt",
			Remote: reduxer.Remote{
				Anonfile: "https://anonfiles.com/naG6S09bu1",
				Catbox:   "https://files.catbox.moe/wwrby6.txt",
			},
		},
		WARC: reduxer.Asset{
			Local: "/path/to/warc",
			Remote: reduxer.Remote{
				Anonfile: "https://anonfiles.com/v4G4S09auc",
				Catbox:   "https://files.catbox.moe/kkai0w.warc",
			},
		},
		Media: reduxer.Asset{
			Local: "",
			Remote: reduxer.Remote{
				Anonfile: "",
				Catbox:   "",
			},
		},
	},
}
