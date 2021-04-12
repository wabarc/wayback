// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"os"
	"testing"

	"github.com/wabarc/wayback/config"
)

func setIRCEnv() {
	os.Setenv("WAYBACK_IRC_NICK", "foo")
	os.Setenv("WAYBACK_IRC_CHANNEL", "bar")

	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()
}

func TestRenderForIRC(t *testing.T) {
	setIRCEnv()

	expected := `Internet Archive:- • https://web.archive.org/web/20211000000001/https://example.com/?q=%E4%B8%AD%E6%96%87, archive.today:- • http://archive.today/abcdE, IPFS:- • https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr, Telegraph:- • http://telegra.ph/title-01-01`

	irc := NewIRC(nil)
	got := irc.Render(collects)
	if got != expected {
		t.Errorf("Unexpected render template for IRC got \n%s\ninstead of \n%s", got, expected)
	}
}

func TestToIRCChannel(t *testing.T) {
}
