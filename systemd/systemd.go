// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package systemd // import "github.com/wabarc/wayback/systemd"

import (
	"net"
	"os"
)

// SdNotifyReady tells the service manager that service startup is
// finished, or the service finished loading its configuration.
// https://www.freedesktop.org/software/systemd/man/sd_notify.html#READY=1
const SdNotifyReady = "READY=1"

// HasNotifySocket checks if the process is supervised by Systemd and has the notify socket.
func HasNotifySocket() bool {
	return os.Getenv("NOTIFY_SOCKET") != ""
}

// SdNotify sends a message to systemd using the sd_notify protocol.
// See https://www.freedesktop.org/software/systemd/man/sd_notify.html.
func SdNotify(state string) error {
	addr := &net.UnixAddr{
		Net:  "unixgram",
		Name: os.Getenv("NOTIFY_SOCKET"),
	}

	// We're not running under systemd (NOTIFY_SOCKET is not set).
	if addr.Name == "" {
		return nil
	}

	conn, err := net.DialUnix(addr.Net, nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err = conn.Write([]byte(state)); err != nil {
		return err
	}

	return nil
}
