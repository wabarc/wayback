// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package config // import "github.com/wabarc/wayback/config"

// Opts holds parsed configuration options.
var Opts *Options

// nolint:stylecheck
const (
	SLOT_IA = "ia"
	SLOT_IS = "is"
	SLOT_IP = "ip"
	SLOT_PH = "ph"
	SLOT_TT = "tt"

	PB_SLUG = "/playback" // Identity for playback
	UNKNOWN = "unknown"
)

// SlotName returns the descriptions of the wayback service.
func SlotName(s string) string {
	slots := map[string]string{
		SLOT_IA: "Internet Archive",
		SLOT_IS: "archive.today",
		SLOT_IP: "IPFS",
		SLOT_PH: "Telegraph",
		SLOT_TT: "Time Travel",
	}

	if _, exist := slots[s]; exist {
		return slots[s]
	}

	return UNKNOWN
}

// SlotExtra returns the extra config of wayback slot.
func SlotExtra(s string) string {
	extra := map[string]string{
		SLOT_IA: "https://web.archive.org/",
		SLOT_IS: "https://archive.today/",
		SLOT_IP: "https://ipfs.github.io/public-gateway-checker/",
		SLOT_PH: "https://telegra.ph/",
		SLOT_TT: "http://timetravel.mementoweb.org/",
	}

	if _, exist := extra[s]; exist {
		return extra[s]
	}

	return UNKNOWN
}
