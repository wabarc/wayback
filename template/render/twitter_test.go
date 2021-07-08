// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"strings"
	"testing"
)

func TestRenderTwitter(t *testing.T) {
	const tweet = `source:
• https://example.com/

————

Internet Archive:
• https://web.archive.org/web/20211000000001/https://example.com/

archive.today:
• http://archive.today/abcdE

IPFS:
• https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr

#wayback #存档`

	got := ForPublish(&Twitter{Cols: collects}).String()
	if got != tweet {
		t.Errorf("Unexpected render template for Twitter got \n%s\ninstead of \n%s", got, tweet)
	}
}

func TestRenderTwitterForReply(t *testing.T) {
	const source = `• https://example.org/`
	const tweet = `Internet Archive:
• https://web.archive.org/123/https://example.com/
• https://web.archive.org/123/https://example.org/

archive.today:
• http://archive.today/abcdE
• http://archive.today/abc

#wayback #存档`
	got := ForReply(&Twitter{Cols: multi}).String()
	if !strings.Contains(got, source) {
		t.Errorf("Unexpected render template for Twitter, got \n%s\ninstead of \n%s", got, source)
	}
	if !strings.Contains(got, tweet) {
		t.Errorf("Unexpected render template for Twitter, got \n%s\ninstead of \n%s", got, tweet)
	}
}
