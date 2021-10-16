// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package systemd // import "github.com/wabarc/wayback/systemd"

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
)

func TestSdNotify(t *testing.T) {
	testDir, err := ioutil.TempDir("", "test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	notifySocket := testDir + "/notify-socket.sock"
	laddr := net.UnixAddr{
		Name: notifySocket,
		Net:  "unixgram",
	}
	if _, err := net.ListenUnixgram("unixgram", &laddr); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		envSocket string

		werr bool
	}{
		// (nil) - notification supported, data has been sent
		{notifySocket, false},
		// (err) - notification supported, but failure happened
		{testDir + "/missing.sock", true},
		// (nil) - notification not supported
		{"", false},
	}

	for i, tt := range tests {
		os.Unsetenv("NOTIFY_SOCKET")

		if tt.envSocket != "" {
			os.Setenv("NOTIFY_SOCKET", tt.envSocket)
		}
		err := SdNotify(fmt.Sprintf("TestSdNotify test message #%d", i))

		if tt.werr && err == nil {
			t.Errorf("#%d: want non-nil err, got nil", i)
		} else if !tt.werr && err != nil {
			t.Errorf("#%d: want nil err, got %v", i, err)
		}
	}
}
