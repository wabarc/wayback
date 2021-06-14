// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package config // import "github.com/wabarc/wayback/config"

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
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

// ParseFile loads configuration values from a local file.
func (p *Parser) ParseFile(filename string) (*Options, error) {
	if filename == "" {
		for _, path := range defaultFilenames() {
			_, err := os.Open(path)
			if err != nil {
				continue
			}
			filename = path
			break
		}
	}

	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	err = p.parseLines(p.parseFileContent(fp))
	if err != nil {
		return nil, err
	}

	return p.opts, nil
}

func (p *Parser) parseFileContent(r io.Reader) (lines []string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "[") && strings.Index(line, "=") > 0 {
			lines = append(lines, line)
		}
	}
	return lines
}

// nolint:gocyclo,unparam
func (p *Parser) parseLines(lines []string) (err error) {
	for _, line := range lines {
		fields := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(fields[0])
		val := strings.TrimSpace(fields[1])

		switch strings.ToUpper(key) {
		case "DEBUG":
			p.opts.debug = parseBool(val, defDebug)
		case "LOG_TIME":
			p.opts.logTime = parseBool(val, defLogTime)
		case "ENABLE_METRICS":
			p.opts.metrics = parseBool(val, defMetrics)
		case "CHROME_REMOTE_ADDR":
			p.opts.enabledChromeRemote = hasValue(val, defEnabledChromeRemote)
			p.opts.chromeRemoteAddr = parseString(val, defChromeRemoteAddr)
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
		case "WAYBACK_TELEGRAM_HELPTEXT":
			p.opts.telegram.helptext = parseString(val, defTelegramHelptext)
		case "WAYBACK_MASTODON_SERVER":
			p.opts.mastodon.server = parseString(val, defMastodonServer)
		case "WAYBACK_MASTODON_KEY":
			p.opts.mastodon.clientKey = parseString(val, defMastodonClientKey)
		case "WAYBACK_MASTODON_SECRET":
			p.opts.mastodon.clientSecret = parseString(val, defMastodonClientSecret)
		case "WAYBACK_MASTODON_TOKEN":
			p.opts.mastodon.accessToken = parseString(val, defMastodonAccessToken)
		case "WAYBACK_TWITTER_CONSUMER_KEY":
			p.opts.twitter.consumerKey = parseString(val, defTwitterConsumerKey)
		case "WAYBACK_TWITTER_CONSUMER_SECRET":
			p.opts.twitter.consumerSecret = parseString(val, defTwitterConsumerSecret)
		case "WAYBACK_TWITTER_ACCESS_TOKEN":
			p.opts.twitter.accessToken = parseString(val, defTwitterAccessToken)
		case "WAYBACK_TWITTER_ACCESS_SECRET":
			p.opts.twitter.accessSecret = parseString(val, defTwitterAccessSecret)
		case "WAYBACK_GITHUB_TOKEN":
			p.opts.github.token = parseString(val, defGitHubToken)
		case "WAYBACK_GITHUB_OWNER":
			p.opts.github.owner = parseString(val, defGitHubOwner)
		case "WAYBACK_GITHUB_REPO":
			p.opts.github.repo = parseString(val, defGitHubRepo)
		case "WAYBACK_IRC_NICK":
			p.opts.irc.nick = parseString(val, defIRCNick)
		case "WAYBACK_IRC_PASSWORD":
			p.opts.irc.password = parseString(val, defIRCPassword)
		case "WAYBACK_IRC_CHANNEL":
			p.opts.irc.channel = parseString(val, defIRCChannel)
		case "WAYBACK_IRC_SERVER":
			p.opts.irc.server = parseString(val, defIRCServer)
		case "WAYBACK_MATRIX_HOMESERVER":
			p.opts.matrix.homeserver = parseString(val, defMatrixHomeserver)
		case "WAYBACK_MATRIX_USERID":
			p.opts.matrix.userID = parseString(val, defMatrixUserID)
		case "WAYBACK_MATRIX_ROOMID":
			p.opts.matrix.roomID = parseString(val, defMatrixRoomID)
		case "WAYBACK_MATRIX_PASSWORD":
			p.opts.matrix.password = parseString(val, defMatrixPassword)
		case "WAYBACK_TOR_PRIVKEY":
			p.opts.tor.pvk = parseString(val, defTorPrivateKey)
		case "WAYBACK_TOR_LOCAL_PORT":
			p.opts.tor.localPort = parseInt(val, defTorLocalPort)
		case "WAYBACK_TOR_REMOTE_PORTS":
			p.opts.tor.remotePorts = parseIntList(val, defTorRemotePorts)
		case "WAYBACK_TORRC":
			p.opts.tor.torrcFile = parseString(val, defTorrcFile)
		case "WAYBACK_POOLING_SIZE":
			p.opts.poolingSize = parseInt(val, defPoolingSize)
		case "WAYBACK_BOLT_PATH":
			p.opts.boltPathname = parseString(val, defBoltPathname)
		case "WAYBACK_STORAGE_DIR":
			p.opts.storageDir = parseString(val, defStorageDir)
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

func hasValue(val string, fallback bool) bool {
	if val == "" {
		return fallback
	}
	return true
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

func defaultFilenames() []string {
	name := "wayback.conf"
	home, _ := os.UserHomeDir()
	return []string{
		name,
		filepath.Join(home, name),
		filepath.Join("/etc", name),
	}
}
