// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"testing"
)

func TestRenderSlack(t *testing.T) {
	message := `‹ Example Domain ›

This domain is for use in illustrative examples in documents. You may use this domain in literature without prior coordination or asking for permission.
More information...

Internet Archive:
• https://web.archive.org/web/20211000000001/https://example.com/

archive.today:
• http://archive.today/abcdE

IPFS:
• https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr

Telegraph:
• http://telegra.ph/title-01-01


<https://anonfiles.com/|AnonFiles> - [ <https://anonfiles.com/FbZfSa9eu4|IMG> ¦ <https://anonfiles.com/r4G8Sb90ud|PDF> ¦ <https://anonfiles.com/pbG4Se94ua|RAW> ¦ <https://anonfiles.com/naG6S09bu1|TXT> ¦ <https://anonfiles.com/n1paZfB3ub|HAR> ¦ <https://anonfiles.com/v4G4S09abc|HTM> ¦ <https://anonfiles.com/v4G4S09auc|WARC> ¦ <|MEDIA> ]
<https://catbox.moe/|Catbox> - [ <https://files.catbox.moe/9u6yvu.png|IMG> ¦ <https://files.catbox.moe/q73uqh.pdf|PDF> ¦ <https://files.catbox.moe/bph1g6.htm|RAW> ¦ <https://files.catbox.moe/wwrby6.txt|TXT> ¦ <https://files.catbox.moe/3agtva.har|HAR> ¦ <|HTM> ¦ <|WARC> ¦ <|MEDIA> ]`

	got := ForPublish(&Slack{Cols: collects, Data: bundleExample}).String()
	if got != message {
		t.Errorf("Unexpected render template for Slack got \n%s\ninstead of \n%s", got, message)
	}
}

func TestRenderSlackFlawed(t *testing.T) {
	message := `Internet Archive:
• Get "https://web.archive.org/save/https://example.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)

archive.today:
• http://archive.today/abcdE

IPFS:
• Archive failed.

Telegraph:
• https://web.archive.org/*/https://webcache.googleusercontent.com/search?q=cache:https://example.com/`

	got := ForPublish(&Slack{Cols: flawed, Data: emptyBundle}).String()
	if got != message {
		t.Errorf("Unexpected render template for Slack, got \n%s\ninstead of \n%s", got, message)
	}
}

func TestRenderSlackForReply(t *testing.T) {
	message := `Internet Archive:
• https://web.archive.org/123/https://example.com/

archive.today:
• http://archive.today/abcdE

Internet Archive:
• https://web.archive.org/123/https://example.org/

archive.today:
• http://archive.today/abc


<https://anonfiles.com/|AnonFiles> - [ <https://anonfiles.com/FbZfSa9eu4|IMG> ¦ <https://anonfiles.com/r4G8Sb90ud|PDF> ¦ <https://anonfiles.com/pbG4Se94ua|RAW> ¦ <https://anonfiles.com/naG6S09bu1|TXT> ¦ <https://anonfiles.com/n1paZfB3ub|HAR> ¦ <https://anonfiles.com/v4G4S09abc|HTM> ¦ <https://anonfiles.com/v4G4S09auc|WARC> ¦ <|MEDIA> ]
<https://catbox.moe/|Catbox> - [ <https://files.catbox.moe/9u6yvu.png|IMG> ¦ <https://files.catbox.moe/q73uqh.pdf|PDF> ¦ <https://files.catbox.moe/bph1g6.htm|RAW> ¦ <https://files.catbox.moe/wwrby6.txt|TXT> ¦ <https://files.catbox.moe/3agtva.har|HAR> ¦ <|HTM> ¦ <|WARC> ¦ <|MEDIA> ]`

	got := ForReply(&Slack{Cols: multi, Data: bundleExample}).String()
	if got != message {
		t.Errorf("Unexpected render template for Slack, got \n%s\ninstead of \n%s", got, message)
	}
}
