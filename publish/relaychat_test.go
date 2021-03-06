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
func TestToIRCChannel(t *testing.T) {
}
