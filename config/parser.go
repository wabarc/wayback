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
	}
	return p.opts, nil
}

// ParseFile loads configuration values from a local file.
func (p *Parser) ParseFile(filename string) (*Options, error) {
	if filename == "" {
		for _, path := range defaultFilenames() {
			_, err := os.Open(filepath.Clean(path))
			if err != nil {
				continue
			}
			filename = path
			break
		}
	}

	fp, err := os.Open(filepath.Clean(filename))
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
// gocyclo:ignore
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
		case "LOG_LEVEL":
			p.opts.logLevel = parseString(val, defLogLevel)
		case "ENABLE_METRICS":
			p.opts.metrics = parseBool(val, defMetrics)
		case "HTTP_LISTEN_ADDR", "WAYBACK_LISTEN_ADDR":
			p.opts.listenAddr = parseString(val, defListenAddr)
		case "CHROME_REMOTE_ADDR":
			p.opts.enabledChromeRemote = hasValue(val, defEnabledChromeRemote)
			p.opts.chromeRemoteAddr = parseString(val, defChromeRemoteAddr)
		case "WAYBACK_IPFS_HOST":
			p.opts.ipfs.host = parseString(val, defIPFSHost)
		case "WAYBACK_IPFS_PORT":
			p.opts.ipfs.port = parseInt(val, defIPFSPort)
		case "WAYBACK_IPFS_MODE":
			p.opts.ipfs.mode = parseString(val, defIPFSMode)
		case "WAYBACK_IPFS_TARGET":
			p.opts.ipfs.target = parseString(val, defIPFSTarget)
		case "WAYBACK_IPFS_APIKEY":
			p.opts.ipfs.apikey = parseString(val, defIPFSApikey)
		case "WAYBACK_IPFS_SECRET":
			p.opts.ipfs.secret = parseString(val, defIPFSSecret)
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
		case "WAYBACK_NOTION_TOKEN":
			p.opts.notion.token = parseString(val, defNotionToken)
		case "WAYBACK_NOTION_DATABASE_ID":
			p.opts.notion.databaseID = parseString(val, defNotionDatabaseID)
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
		case "WAYBACK_DISCORD_BOT_TOKEN":
			p.opts.discord.botToken = parseString(val, defDiscordBotToken)
		case "WAYBACK_DISCORD_CHANNEL":
			p.opts.discord.channel = parseString(val, defDiscordChannel)
		case "WAYBACK_DISCORD_HELPTEXT":
			p.opts.discord.helptext = parseString(val, defDiscordHelptext)
		case "WAYBACK_SLACK_APP_TOKEN":
			p.opts.slack.appToken = parseString(val, defSlackAppToken)
		case "WAYBACK_SLACK_BOT_TOKEN":
			p.opts.slack.botToken = parseString(val, defSlackBotToken)
		case "WAYBACK_SLACK_CHANNEL":
			p.opts.slack.channel = parseString(val, defSlackChannel)
		case "WAYBACK_SLACK_HELPTEXT":
			p.opts.slack.helptext = parseString(val, defSlackHelptext)
		case "WAYBACK_NOSTR_RELAY_URL":
			p.opts.nostr.url = parseString(val, defNostrRelayURL)
		case "WAYBACK_NOSTR_PRIVATE_KEY":
			p.opts.nostr.privateKey = parseString(val, defNostrPrivateKey)
		case "WAYBACK_TOR_PRIVKEY", "WAYBACK_ONION_PRIVKEY":
			p.opts.onion.pvk = parseString(val, defOnionPrivateKey)
		case "WAYBACK_TOR_LOCAL_PORT", "WAYBACK_ONION_LOCAL_PORT":
			p.opts.onion.localPort = parseInt(val, defOnionLocalPort)
		case "WAYBACK_TOR_REMOTE_PORTS", "WAYBACK_ONION_REMOTE_PORTS":
			p.opts.onion.remotePorts = parseIntList(val, defOnionRemotePorts)
		case "WAYBACK_ONION_DISABLED":
			p.opts.onion.disabled = parseBool(val, defOnionDisabled)
		case "WAYBACK_POOLING_SIZE":
			p.opts.poolingSize = parseInt(val, defPoolingSize)
		case "WAYBACK_BOLT_PATH":
			p.opts.boltPathname = parseString(val, defBoltPathname)
		case "WAYBACK_STORAGE_DIR":
			p.opts.storageDir = parseString(val, defStorageDir)
		case "WAYBACK_MAX_MEDIA_SIZE":
			p.opts.maxMediaSize = parseString(val, defMaxMediaSize)
		case "WAYBACK_TIMEOUT":
			p.opts.waybackTimeout = parseInt(val, defWaybackTimeout)
		case "WAYBACK_MAX_RETRIES":
			p.opts.waybackMaxRetries = parseInt(val, defWaybackMaxRetries)
		case "WAYBACK_USERAGENT":
			p.opts.waybackUserAgent = parseString(val, defWaybackUserAgent)
		case "WAYBACK_FALLBACK":
			p.opts.waybackFallback = parseBool(val, defWaybackFallback)
		case "WAYBACK_MEILI_ENDPOINT":
			p.opts.waybackMeiliEndpoint = parseString(val, defWaybackMeiliEndpoint)
		case "WAYBACK_MEILI_INDEXING":
			p.opts.waybackMeiliIndexing = parseString(val, defWaybackMeiliIndexing)
		case "WAYBACK_MEILI_APIKEY":
			p.opts.waybackMeiliApikey = parseString(val, defWaybackMeiliApikey)
		default:
			if os.Getenv(key) == "" && val != "" {
				os.Setenv(key, val)
			}
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

	items := strings.Split(val, ",")
	intList := make([]int, 0, len(items))
	for _, item := range items {
		i, _ := strconv.Atoi(strings.TrimSpace(item)) // nolint:errcheck
		intList = append(intList, i)
	}

	return intList
}

func defaultFilenames() []string {
	name := "wayback.conf"
	home, _ := os.UserHomeDir() // nolint:errcheck
	return []string{
		name,
		filepath.Join(home, name),
		filepath.Join("/", "etc", name),
	}
}

func breakLine(s string) string {
	s = strings.ReplaceAll(s, `\r`, "\n")
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\r\n`, "\n")
	s = strings.ReplaceAll(s, `<br>`, "\n")
	s = strings.ReplaceAll(s, `<br/>`, "\n")
	return s
}
