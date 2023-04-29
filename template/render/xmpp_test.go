// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"testing"
)

func TestRenderForXMPP(t *testing.T) {
	expected := `Internet Archive:
• https://web.archive.org/web/20211000000001/https://example.com/

IPFS:
• https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr

archive.today:
• http://archive.today/abcdE

Telegraph:
• http://telegra.ph/title-01-01`

	got := ForPublish(&XMPP{Cols: collects, Data: bundleExample}).String()
	if got != expected {
		t.Errorf("Unexpected render template for XMPP, got \n%s\ninstead of \n%s", got, expected)
	}
}
