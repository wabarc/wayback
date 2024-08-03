// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"testing"

	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
)

func TestRenderGitHub(t *testing.T) {
	collects := []wayback.Collect{
		{
			Arc: config.SLOT_IA,
			Dst: "https://web.archive.org/web/20211000000001/https://example.com/?q=%E4%B8%AD%E6%96%87",
			Src: "https://example.com/?q=%E4%B8%AD%E6%96%87",
			Ext: config.SLOT_IA,
		},
		{
			Arc: config.SLOT_IS,
			Dst: "http://archive.today/abcdE",
			Src: "https://example.com/?q=%E4%B8%AD%E6%96%87",
			Ext: config.SLOT_IS,
		},
		{
			Arc: config.SLOT_IP,
			Dst: "https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr",
			Src: "https://example.com/?q=%E4%B8%AD%E6%96%87",
			Ext: config.SLOT_IP,
		},
		{
			Arc: config.SLOT_PH,
			Dst: "http://telegra.ph/title-01-01",
			Src: "https://example.com/?q=%E4%B8%AD%E6%96%87",
			Ext: config.SLOT_PH,
		},
	}

	expected := `**[Internet Archive](https://web.archive.org/)**:
> source: [https://example.com/?q=中文](https://example.com/?q=%E4%B8%AD%E6%96%87)
> archived: [https://web.archive.org/web/20211000000001/https://example.com/?q=中文](https://web.archive.org/web/20211000000001/https://example.com/?q=%E4%B8%AD%E6%96%87)

**[archive.today](https://archive.today/)**:
> source: [https://example.com/?q=中文](https://example.com/?q=%E4%B8%AD%E6%96%87)
> archived: [http://archive.today/abcdE](http://archive.today/abcdE)

**[IPFS](https://ipfs.github.io/public-gateway-checker/)**:
> source: [https://example.com/?q=中文](https://example.com/?q=%E4%B8%AD%E6%96%87)
> archived: [https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr](https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr)

**[Telegraph](https://telegra.ph/)**:
> source: [https://example.com/?q=中文](https://example.com/?q=%E4%B8%AD%E6%96%87)
> archived: [http://telegra.ph/title-01-01](http://telegra.ph/title-01-01)`

	got := ForPublish(&GitHub{Cols: collects, Data: bundleExample}).String()
	if got != expected {
		t.Errorf("Unexpected render template for GitHub Issues, got \n%s\ninstead of \n%s", got, expected)
	}
}

func TestRenderGitHubFlawed(t *testing.T) {
	expected := `**[Internet Archive](https://web.archive.org/)**:
> source: [https://example.com/](https://example.com/)
> archived: Get "https://web.archive.org/save/https://example.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)

**[archive.today](https://archive.today/)**:
> source: [https://example.com/](https://example.com/)
> archived: [http://archive.today/abcdE](http://archive.today/abcdE)

**[IPFS](https://ipfs.github.io/public-gateway-checker/)**:
> source: [https://example.com/](https://example.com/)
> archived: Archive failed.

**[Telegraph](https://telegra.ph/)**:
> source: [https://example.com/](https://example.com/)
> archived: [https://web.archive.org/*/https://webcache.googleusercontent.com/search?q=cache:https://example.com/](https://web.archive.org/*/https://webcache.googleusercontent.com/search?q=cache:https://example.com/)`

	got := ForPublish(&GitHub{Cols: flawed, Data: emptyBundle}).String()
	if got != expected {
		t.Errorf("Unexpected render template for GitHub Issues, got \n%s\ninstead of \n%s", got, expected)
	}
}

func TestRenderGitHubWithArtifact(t *testing.T) {
	const expected = `This domain is for use in illustrative examples in documents. You may use this domain in literature without prior coordination or asking for permission.

More information...

**[Internet Archive](https://web.archive.org/)**:
> source: [https://example.com/](https://example.com/)
> archived: [https://web.archive.org/web/20211000000001/https://example.com/](https://web.archive.org/web/20211000000001/https://example.com/)

**[archive.today](https://archive.today/)**:
> source: [https://example.com/](https://example.com/)
> archived: [http://archive.today/abcdE](http://archive.today/abcdE)

**[IPFS](https://ipfs.github.io/public-gateway-checker/)**:
> source: [https://example.com/](https://example.com/)
> archived: [https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr](https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr)

**[Telegraph](https://telegra.ph/)**:
> source: [https://example.com/](https://example.com/)
> archived: [http://telegra.ph/title-01-01](http://telegra.ph/title-01-01)

**[Catbox](https://catbox.moe/)** - [ [IMG](https://files.catbox.moe/9u6yvu.png) ¦ [PDF](https://files.catbox.moe/q73uqh.pdf) ¦ [RAW](https://files.catbox.moe/bph1g6.htm) ¦ [TXT](https://files.catbox.moe/wwrby6.txt) ¦ [HAR](https://files.catbox.moe/3agtva.har) ¦ [HTM]() ¦ [WARC]() ¦ [MEDIA]() ]`

	got := ForPublish(&GitHub{Cols: collects, Data: bundleExample}).String()
	if got != expected {
		t.Errorf("Unexpected render template for Matrix, got \n%s\ninstead of \n%s", got, expected)
	}
}
