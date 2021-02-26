// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package config // import "github.com/wabarc/wayback/config"

import (
	"os"
	"strconv"
	"strings"
)

// Parser handles configuration parsing.
type Parser struct {
	opts *Options
}

// NewParser returns a new Parser.
func NewParser() *Parser {
	return &Parser{
		opts: NewOptions(),
	}
}

// ParseEnvironmentVariables loads configuration values from environment variables.
func (p *Parser) ParseEnvironmentVariables() (*Options, error) {
	if err := p.parseLines(os.Environ()); err != nil {
		return nil, err
	} else {
		return p.opts, nil
	}
}

func (p *Parser) parseLines(lines []string) (err error) {
	for _, line := range lines {
		fields := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(fields[0])
		val := strings.TrimSpace(fields[1])

		switch key {
		case "DEBUG":
			p.opts.debug = parseBool(val, defDebug)
		case "LOG_TIME":
			p.opts.logTime = parseBool(val, defLogTime)
		case "WAYBACK_IPFS_HOST":
			p.opts.ipfs.host = parseString(val, defIPFSHost)
		case "WAYBACK_IPFS_PORT":
			p.opts.ipfs.port = uint(parseInt(val, defIPFSPort))
		case "WAYBACK_IPFS_MODE":
			p.opts.ipfs.mode = parseString(val, defIPFSMode)
		case "WAYBACK_USE_TOR":
			p.opts.overTor = parseBool(val, defOverTor)
		case "WAYBACK_ENABLE_IA":
			p.opts.slots[SLOT_IA] = parseBool(val, defEnabledIA)
		case "WAYBACK_ENABLE_IS":
			p.opts.slots[SLOT_IS] = parseBool(val, defEnabledIS)
		case "WAYBACK_ENABLE_IP":
			p.opts.slots[SLOT_IP] = parseBool(val, defEnabledIP)
		case "WAYBACK_ENABLE_PH":
			p.opts.slots[SLOT_PH] = parseBool(val, defEnabledPH)
		case "WAYBACK_TELEGRAM_TOKEN":
			p.opts.telegram.token = parseString(val, defTelegramToken)
		case "WAYBACK_TELEGRAM_CHANNEL":
			p.opts.telegram.channel = parseString(val, defTelegramChannel)
		case "WAYBACK_MASTODON_SERVER":
			p.opts.mastodon.server = parseString(val, defMastodonServer)
		case "WAYBACK_MASTODON_KEY":
			p.opts.mastodon.clientKey = parseString(val, defMastodonClientKey)
		case "WAYBACK_MASTODON_SECRET":
			p.opts.mastodon.clientSecret = parseString(val, defMastodonClientSecret)
		case "WAYBACK_MASTODON_TOKEN":
			p.opts.mastodon.accessToken = parseString(val, defMastodonAccessToken)
		case "WAYBACK_GITHUB_TOKEN":
			p.opts.github.token = parseString(val, defGitHubToken)
		case "WAYBACK_GITHUB_OWNER":
			p.opts.github.owner = parseString(val, defGitHubOwner)
		case "WAYBACK_GITHUB_REPO":
			p.opts.github.repo = parseString(val, defGitHubRepo)
		case "WAYBACK_TOR_PRIVKEY":
			p.opts.tor.pvk = parseString(val, defTorPrivateKey)
		case "WAYBACK_TOR_LOCAL_PORT":
			p.opts.tor.localPort = parseInt(val, defTorLocalPort)
		case "WAYBACK_TOR_REMOTE_PORTS":
			p.opts.tor.remotePorts = parseIntList(val, defTorRemotePorts)
		case "WAYBACK_TORRC":
			p.opts.tor.torrcFile = parseString(val, defTorrcFile)
		}
	}

	return nil
}

func parseBool(val string, fallback bool) bool {
	if val == "" {
		return fallback
	}

	val = strings.ToLower(val)
	if val == "1" || val == "yes" || val == "true" || val == "on" {
		return true
	}

	return false
}

func parseInt(val string, fallback int) int {
	if val == "" {
		return fallback
	}

	v, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return v
}

func parseString(val string, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}

func parseIntList(val string, fallback []int) []int {
	if val == "" {
		return fallback
	}

	var intList []int
	items := strings.Split(val, ",")
	for _, item := range items {
		i, _ := strconv.Atoi(strings.TrimSpace(item))
		intList = append(intList, i)
	}

	return intList
}
