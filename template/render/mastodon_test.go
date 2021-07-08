// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

/*
Package render handles template parsing and execution for services.
*/

package render // import "github.com/wabarc/wayback/template/render"

import (
	"strings"
	"testing"
)

func TestRenderMastodon(t *testing.T) {
	const source = `source:
• https://example.com/

————

`
	const toot = `Internet Archive:
• https://web.archive.org/web/20211000000001/https://example.com/

archive.today:
• http://archive.today/abcdE

IPFS:
• https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr

Telegraph:
• http://telegra.ph/title-01-01

#wayback #存档`

	got := ForPublish(&Mastodon{Cols: collects}).String()
	if !strings.Contains(got, source) {
		t.Fatalf("Unexpected render template for Mastodon, got \n%s\ninstead of \n%s", got, source)
	}
	if !strings.Contains(got, toot) {
		t.Fatalf("Unexpected render template for Mastodon, got \n%s\ninstead of \n%s", got, toot)
	}
}
