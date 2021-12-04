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
• <a href="https://example.com/">source</a> - http://telegra.ph/title-01-01<br>
<br>
<b><a href="https://anonfiles.com/">AnonFiles</a></b> - [ <a href="">IMG</a> ¦ <a href="">PDF</a> ¦ <a href="">RAW</a> ¦ <a href="">TXT</a> ¦ <a href="">HAR</a> ¦ <a href="">WARC</a> ¦ <a href="">MEDIA</a> ]<br>
<b><a href="https://catbox.moe/">Catbox</a></b> - [ <a href="">IMG</a> ¦ <a href="">PDF</a> ¦ <a href="">RAW</a> ¦ <a href="">TXT</a> ¦ <a href="">HAR</a> ¦ <a href="">WARC</a> ¦ <a href="">MEDIA</a> ]`

	got := ForPublish(&Matrix{Cols: collects}).String()
	if got != matExp {
		t.Errorf("Unexpected render template for Matrix, got \n%s\ninstead of \n%s", got, matExp)
	}
}

func TestRenderMatrixWithAssets(t *testing.T) {
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
• <a href="https://example.com/">source</a> - http://telegra.ph/title-01-01<br>
<br>
<b><a href="https://anonfiles.com/">AnonFiles</a></b> - [ <a href="https://anonfiles.com/FbZfSa9eu4">IMG</a> ¦ <a href="https://anonfiles.com/r4G8Sb90ud">PDF</a> ¦ <a href="https://anonfiles.com/pbG4Se94ua">RAW</a> ¦ <a href="https://anonfiles.com/naG6S09bu1">TXT</a> ¦ <a href="https://anonfiles.com/n1paZfB3ub">HAR</a> ¦ <a href="https://anonfiles.com/v4G4S09auc">WARC</a> ¦ <a href="">MEDIA</a> ]<br>
<b><a href="https://catbox.moe/">Catbox</a></b> - [ <a href="https://files.catbox.moe/9u6yvu.png">IMG</a> ¦ <a href="https://files.catbox.moe/q73uqh.pdf">PDF</a> ¦ <a href="https://files.catbox.moe/bph1g6.htm">RAW</a> ¦ <a href="https://files.catbox.moe/wwrby6.txt">TXT</a> ¦ <a href="https://files.catbox.moe/3agtva.har">HAR</a> ¦ <a href="">WARC</a> ¦ <a href="">MEDIA</a> ]`

	got := ForPublish(&Matrix{Cols: collects, Data: bundleExample}).String()
	if got != matExp {
		t.Errorf("Unexpected render template for Matrix, got \n%s\ninstead of \n%s", got, matExp)
	}
}
