// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"testing"
)

func TestRenderNostrForReply(t *testing.T) {
	TestRenderNostrForPublish(t)
}

func TestRenderNostrForPublish(t *testing.T) {
	expected := `‹ Example ›

source:
• https://example.com/

————

Internet Archive:
• https://web.archive.org/web/20211000000001/https://example.com/
archive.today:
• http://archive.today/abcdE
IPFS:
• https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr
Telegraph:
• http://telegra.ph/title-01-01`

	got := ForPublish(&Nostr{Cols: collects, Data: bundleExample}).String()
	if got != expected {
		t.Errorf("Unexpected render template for IRC, got \n%s\ninstead of \n%s", got, expected)
	}
}
