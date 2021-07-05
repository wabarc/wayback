// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"testing"
)

func TestRenderMatrix(t *testing.T) {
	const matExp = `<b><a href='https://web.archive.org/'>Internet Archive</a></b>:<br>
• <a href="https://example.com/">source</a> - https://web.archive.org/web/20211000000001/https://example.com/<br>
<br>
<b><a href='https://archive.today/'>archive.today</a></b>:<br>
• <a href="https://example.com/">source</a> - http://archive.today/abcdE<br>
<br>
<b><a href='https://ipfs.github.io/public-gateway-checker/'>IPFS</a></b>:<br>
• <a href="https://example.com/">source</a> - https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr<br>
<br>
<b><a href='https://telegra.ph/'>Telegraph</a></b>:<br>
• <a href="https://example.com/">source</a> - http://telegra.ph/title-01-01<br>`

	got := ForPublish(&Matrix{Cols: collects}).String()
	if got != matExp {
		t.Errorf("Unexpected render template for Matrix, got \n%s\ninstead of \n%s", got, matExp)
	}
}
