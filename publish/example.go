// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
)

var Collects = []wayback.Collect{
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
