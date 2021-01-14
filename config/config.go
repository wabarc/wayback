// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package config // import "github.com/wabarc/wayback/config"

// Opts holds parsed configuration options.
var Opts *Options

const (
	SLOT_IA = "ia"
	SLOT_IS = "is"
	SLOT_IP = "ip"
)

// SlotName returns the descriptions of the wayback service.
func SlotName(s string) string {
	slots := map[string]string{
		SLOT_IA: "Internet Archive",
		SLOT_IS: "archive.today",
		SLOT_IP: "IPFS",
	}

	return slots[s]
}
