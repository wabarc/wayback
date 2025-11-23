// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package config // import "github.com/wabarc/wayback/config"

// Opts holds parsed configuration options.
// var Opts *Options

const (
	SLOT_IA = "ia" // Internet Archive
	SLOT_IS = "is" // archive.today
	SLOT_IP = "ip" // IPFS
	SLOT_PH = "ph" // Telegraph
	SLOT_GA = "ga" // Ghostarchive
	SLOT_TT = "tt" // Time Travel

	PB_SLUG = "/playback" // Identity for playback
	UNKNOWN = "unknown"
)

const (
	ServiceDiscord  Flag = iota + 1 // FlagDiscord represents discord service
	ServiceHTTPd                    // FlagWeb represents httpd service
	ServiceMastodon                 // FlagMastodon represents mastodon service
	ServiceMatrix                   // FlagMatrix represents matrix service
	ServiceIRC                      // FlagIRC represents relaychat service
	ServiceSlack                    // FlagSlack represents slack service
	ServiceTelegram                 // FlagTelegram represents telegram service
	ServiceTwitter                  // FlagTwitter represents twitter srvice
	ServiceXMPP                     // FlagXMPP represents XMPP service
)

// Flag represents a type of uint8
type Flag uint8

// String returns the flag as a string.
func (f Flag) String() string {
	switch f {
	case ServiceHTTPd:
		return "httpd"
	case ServiceTelegram:
		return "telegram"
	case ServiceTwitter:
		return "twiter"
	case ServiceMastodon:
		return "mastodon"
	case ServiceDiscord:
		return "discord"
	case ServiceMatrix:
		return "matrix"
	case ServiceSlack:
		return "slack"
	case ServiceIRC:
		return "relaychat"
	case ServiceXMPP:
		return "xmpp"
	default:
		return ""
	}
}

// SlotName returns the descriptions of the wayback service.
func SlotName(s string) string {
	slots := map[string]string{
		SLOT_IA: "Internet Archive",
		SLOT_IS: "archive.today",
		SLOT_IP: "IPFS",
		SLOT_PH: "Telegraph",
		SLOT_GA: "Ghostarchive",
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
		SLOT_GA: "https://ghostarchive.org/",
		SLOT_TT: "http://timetravel.mementoweb.org/",
	}

	if _, exist := extra[s]; exist {
		return extra[s]
	}

	return UNKNOWN
}
