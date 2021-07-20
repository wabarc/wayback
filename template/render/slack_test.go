// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"testing"
)

func TestRenderSlack(t *testing.T) {
	message := `Internet Archive:
• https://web.archive.org/web/20211000000001/https://example.com/

archive.today:
• http://archive.today/abcdE

IPFS:
• https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr

Telegraph:
• http://telegra.ph/title-01-01`
	got := ForPublish(&Slack{Cols: collects}).String()
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
• https://web.archive.org/*/https://webcache.googleusercontent.com/search?q=cache:https://example.com/404`
	got := ForPublish(&Slack{Cols: flawed}).String()
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
• http://archive.today/abc`
	got := ForReply(&Slack{Cols: multi}).String()
	if got != message {
		t.Errorf("Unexpected render template for Slack, got \n%s\ninstead of \n%s", got, message)
	}
}
