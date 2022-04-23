// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"testing"

	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
)

func TestRenderNotion(t *testing.T) {
	collects := []wayback.Collect{
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

	expected := `<!doctype html>
<html>
<head>
    <title>Example Domain</title>
</head>

<body>
<div>
    <h1>Example Domain</h1>
    <p>This domain is for use in illustrative examples in documents. You may use this
    domain in literature without prior coordination or asking for permission.</p>
    <p><a href="https://www.iana.org/domains/example">More information...</a></p>
</div>
</body>
</html>`

	got := ForPublish(&Notion{Cols: collects, Data: bundleExample}).String()
	if got != expected {
		t.Errorf("Unexpected render template for Notion Issues, got \n%s\ninstead of \n%s", got, expected)
	}
}
