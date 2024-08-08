// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"testing"
)

var message = `<b>Example</b>

This domain is for use in illustrative examples in documents. You may use this domain in literature without prior coordination or asking for permission.

More information...


<b><a href="https://web.archive.org/">Internet Archive</a></b>:
• <a href="https://example.com/">source</a> - <a href="https://web.archive.org/web/20211000000001/https://example.com/">https://web.archive.org/web/20211000000001/https://example.com/</a>

<b><a href="https://archive.today/">archive.today</a></b>:
• <a href="https://example.com/">source</a> - <a href="http://archive.today/abcdE">http://archive.today/abcdE</a>

<b><a href="https://ipfs.github.io/public-gateway-checker/">IPFS</a></b>:
• <a href="https://example.com/">source</a> - <a href="https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr">https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr</a>

<b><a href="https://telegra.ph/">Telegraph</a></b>:
• <a href="https://example.com/">source</a> - <a href="http://telegra.ph/title-01-01">http://telegra.ph/title-01-01</a>
`

func TestRenderTelegram(t *testing.T) {
	message := message + `
<b><a href="https://catbox.moe/">Catbox</a></b> - [ <a href="https://files.catbox.moe/9u6yvu.png">IMG</a> ¦ <a href="https://files.catbox.moe/q73uqh.pdf">PDF</a> ¦ <a href="https://files.catbox.moe/bph1g6.htm">RAW</a> ¦ <a href="https://files.catbox.moe/wwrby6.txt">TXT</a> ¦ <a href="https://files.catbox.moe/3agtva.har">HAR</a> ¦ <a href="">HTM</a> ¦ <a href="">WARC</a> ¦ <a href="">MEDIA</a> ]

#wayback #存档`

	got := ForPublish(&Telegram{Cols: collects, Data: bundleExample}).String()
	if got != message {
		t.Errorf("Unexpected render template for Telegram got \n%s\ninstead of \n%s", got, message)
	}
}

func TestRenderTelegramForPublishWithArtifact(t *testing.T) {
	message := message + `
<b><a href="https://catbox.moe/">Catbox</a></b> - [ <a href="https://files.catbox.moe/9u6yvu.png">IMG</a> ¦ <a href="https://files.catbox.moe/q73uqh.pdf">PDF</a> ¦ <a href="https://files.catbox.moe/bph1g6.htm">RAW</a> ¦ <a href="https://files.catbox.moe/wwrby6.txt">TXT</a> ¦ <a href="https://files.catbox.moe/3agtva.har">HAR</a> ¦ <a href="">HTM</a> ¦ <a href="">WARC</a> ¦ <a href="">MEDIA</a> ]

#wayback #存档`

	got := ForPublish(&Telegram{Cols: collects, Data: bundleExample}).String()
	if got != message {
		t.Errorf("Unexpected render template for Telegram got \n%s\ninstead of \n%s", got, message)
	}
}

func TestRenderTelegramFlawed(t *testing.T) {
	message := `<b><a href="https://web.archive.org/">Internet Archive</a></b>:
• <a href="https://example.com/">source</a> - Get &#34;https://web.archive.org/save/https://example.com&#34;: context deadline exceeded (Client.Timeout exceeded while awaiting headers)

<b><a href="https://archive.today/">archive.today</a></b>:
• <a href="https://example.com/">source</a> - <a href="http://archive.today/abcdE">http://archive.today/abcdE</a>

<b><a href="https://ipfs.github.io/public-gateway-checker/">IPFS</a></b>:
• <a href="https://example.com/">source</a> - Archive failed.

<b><a href="https://telegra.ph/">Telegraph</a></b>:
• <a href="https://example.com/">source</a> - <a href="https://web.archive.org/*/https://webcache.googleusercontent.com/search?q=cache:https://example.com/">https://web.archive.org/*/https://webcache.googleusercontent.com/search?q=cache:https://example.com/</a>

#wayback #存档`

	got := ForPublish(&Telegram{Cols: flawed, Data: emptyBundle}).String()
	if got != message {
		t.Errorf("Unexpected render template for Telegram, got \n%s\ninstead of \n%s", got, message)
	}
}

func TestRenderTelegramForReply(t *testing.T) {
	message := `<b><a href="https://web.archive.org/">Internet Archive</a></b>:
• <a href="https://example.com/">source</a> - <a href="https://web.archive.org/123/https://example.com/">https://web.archive.org/123/https://example.com/</a>
• <a href="https://example.org/">source</a> - <a href="https://web.archive.org/123/https://example.org/">https://web.archive.org/123/https://example.org/</a>

<b><a href="https://archive.today/">archive.today</a></b>:
• <a href="https://example.com/">source</a> - <a href="http://archive.today/abcdE">http://archive.today/abcdE</a>
• <a href="https://example.org/">source</a> - <a href="http://archive.today/abc">http://archive.today/abc</a>

<b><a href="https://catbox.moe/">Catbox</a></b> - [ <a href="https://files.catbox.moe/9u6yvu.png">IMG</a> ¦ <a href="https://files.catbox.moe/q73uqh.pdf">PDF</a> ¦ <a href="https://files.catbox.moe/bph1g6.htm">RAW</a> ¦ <a href="https://files.catbox.moe/wwrby6.txt">TXT</a> ¦ <a href="https://files.catbox.moe/3agtva.har">HAR</a> ¦ <a href="">HTM</a> ¦ <a href="">WARC</a> ¦ <a href="">MEDIA</a> ]

#wayback #存档`

	got := ForReply(&Telegram{Cols: multi, Data: bundleExample}).String()
	if got != message {
		t.Errorf("Unexpected render template for Telegram, got \n%s\ninstead of \n%s", got, message)
	}
}
