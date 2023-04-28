// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package relaychat // import "github.com/wabarc/wayback/publish/relaychat"

import (
	"os"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

const server = "irc.libera.chat:6697"

func setIRCEnv() {
	os.Setenv("WAYBACK_IRC_NICK", helper.RandString(6, ""))
	os.Setenv("WAYBACK_IRC_CHANNEL", "bar")
	os.Setenv("WAYBACK_IRC_SERVER", server)
}
func TestToIRCChannel(t *testing.T) {
}

func TestShutdown(t *testing.T) {
	setIRCEnv()
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	irc := New(nil, opts)
	err := irc.Shutdown()
	if err != nil {
		t.Errorf("Unexpected shutdown: %v", err)
	}
}
