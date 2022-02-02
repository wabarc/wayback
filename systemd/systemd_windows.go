// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package systemd // import "github.com/wabarc/wayback/systemd"

const SdNotifyReady = "READY=1"

func HasNotifySocket() bool {
	return false
}

func SdNotify(state string) error {
	return nil
}
