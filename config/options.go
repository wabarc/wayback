// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package config // import "github.com/wabarc/wayback/config"

import (
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/wabarc/logger"
)

const (
	defDebug    = false
	defLogTime  = true
	defLogLevel = "info"
	defMetrics  = false
	defOverTor  = false

	defIPFSHost   = "127.0.0.1"
	defIPFSPort   = 5001
	defIPFSMode   = "pinner"
	defIPFSTarget = ""
	defIPFSApikey = ""
	defIPFSSecret = ""

	defEnabledIA = true
	defEnabledIS = true
	defEnabledIP = true
	defEnabledPH = true

	defTelegramToken    = ""
	defTelegramChannel  = ""
	defTelegramHelptext = "Hi there."
	defGitHubToken      = ""
	defGitHubOwner      = ""
	defGitHubRepo       = ""
	defNotionToken      = ""
	defNotionDatabaseID = ""

	defMastodonServer        = "https://mastodon.social"
	defMastodonClientKey     = ""
	defMastodonClientSecret  = ""
	defMastodonAccessToken   = ""
	defTwitterConsumerKey    = ""
	defTwitterConsumerSecret = ""
	defTwitterAccessToken    = ""
	defTwitterAccessSecret   = ""
	defDiscordBotToken       = ""
	defDiscordChannel        = ""
	defDiscordHelptext       = "Hi there."
	defSlackAppToken         = ""
	defSlackBotToken         = ""
	defSlackChannel          = ""
	defSlackHelptext         = "Hi there."

	defIRCNick     = ""
	defIRCPassword = ""
	defIRCChannel  = ""
	defIRCServer   = "irc.libera.chat:6697"

	defXMPPUsername = ""
	defXMPPPassword = ""
	defXMPPNoTLS    = false
	defXMPPHelptext = "Hi there."

	defMatrixHomeserver = "https://matrix.org"
	defMatrixUserID     = ""
	defMatrixRoomID     = ""
	defMatrixPassword   = ""

	defNostrRelayURL   = "wss://nostr.developer.li"
	defNostrPrivateKey = ""

	defListenAddr      = "0.0.0.0:8964"
	defOnionLocalPort  = 8964
	defOnionPrivateKey = ""
	defOnionDisabled   = false

	defChromeRemoteAddr    = ""
	defEnabledChromeRemote = false
	defBoltPathname        = "wayback.db"
	defPoolingSize         = 3
	defMaxMediaSize        = "512MB"
	defWaybackTimeout      = 300
	defWaybackMaxRetries   = 2
	defWaybackUserAgent    = "WaybackArchiver/1.0"
	defWaybackFallback     = false
	defProxy               = ""

	defWaybackMeiliEndpoint = ""
	defWaybackMeiliIndexing = "capsules"
	defWaybackMeiliApikey   = ""

	maxAttachSizeTelegram = 50000000   // 50MB
	maxAttachSizeDiscord  = 8000000    // 8MB
	maxAttachSizeSlack    = 5000000000 // 5GB
)

var (
	IPFSToken  = ""
	IPFSTarget = "web3storage"

	defStorageDir       = path.Join(os.TempDir(), "reduxer")
	defOnionRemotePorts = []int{80}
)

// Options represents a configuration options in the application.
type Options struct {
	debug    bool
	logTime  bool
	logLevel string
	overTor  bool
	metrics  bool

	// Enabled services
	services sync.Map

	ipfs     *ipfs
	slots    map[string]bool
	telegram *telegram
	mastodon *mastodon
	discord  *discord
	twitter  *twitter
	github   *github
	notion   *notion
	matrix   *matrix
	slack    *slack
	nostr    *nostr
	irc      *irc
	onion    *onion
	xmpp     *xmpp

	listenAddr          string
	chromeRemoteAddr    string
	enabledChromeRemote bool
	boltPathname        string
	poolingSize         int
	storageDir          string
	maxMediaSize        string
	proxy               string
	waybackTimeout      int
	waybackMaxRetries   int
	waybackUserAgent    string
	waybackFallback     bool

	waybackMeiliEndpoint string
	waybackMeiliIndexing string
	waybackMeiliApikey   string
}

type ipfs struct {
	host string
	port int
	mode string

	target string
	apikey string
	secret string
}

type telegram struct {
	token    string
	channel  string
	helptext string
}

type mastodon struct {
	server       string
	clientKey    string
	clientSecret string
	accessToken  string
}

type discord struct {
	appID    string
	botToken string
	channel  string
	helptext string
}

type twitter struct {
	consumerKey    string
	consumerSecret string
	accessToken    string
	accessSecret   string
}

type github struct {
	token string
	owner string
	repo  string
}

type notion struct {
	token      string
	databaseID string
}

type matrix struct {
	homeserver string
	userID     string
	roomID     string
	password   string
}

type slack struct {
	appToken string
	botToken string
	channel  string
	helptext string
}

type nostr struct {
	url        string
	privateKey string
}

type irc struct {
	nick     string
	password string
	channel  string
	server   string
}

type onion struct {
	pvk string

	localPort   int
	remotePorts []int

	disabled bool
}

type xmpp struct {
	username string
	password string
	noTLS    bool
	helptext string
}

// NewOptions returns Options with default values.
func NewOptions() *Options {
	opts := &Options{
		debug:                defDebug,
		logTime:              defLogTime,
		logLevel:             defLogLevel,
		overTor:              defOverTor,
		metrics:              defMetrics,
		listenAddr:           defListenAddr,
		chromeRemoteAddr:     defChromeRemoteAddr,
		enabledChromeRemote:  defEnabledChromeRemote,
		boltPathname:         defBoltPathname,
		poolingSize:          defPoolingSize,
		storageDir:           defStorageDir,
		maxMediaSize:         defMaxMediaSize,
		waybackTimeout:       defWaybackTimeout,
		waybackMaxRetries:    defWaybackMaxRetries,
		waybackUserAgent:     defWaybackUserAgent,
		waybackFallback:      defWaybackFallback,
		waybackMeiliEndpoint: defWaybackMeiliEndpoint,
		waybackMeiliIndexing: defWaybackMeiliIndexing,
		waybackMeiliApikey:   defWaybackMeiliApikey,
		ipfs: &ipfs{
			host:   defIPFSHost,
			port:   defIPFSPort,
			mode:   defIPFSMode,
			target: defIPFSTarget,
			apikey: defIPFSApikey,
			secret: defIPFSSecret,
		},
		slots: map[string]bool{
			SLOT_IA: defEnabledIA,
			SLOT_IS: defEnabledIS,
			SLOT_IP: defEnabledIP,
			SLOT_PH: defEnabledPH,
		},
		telegram: &telegram{
			token:    defTelegramToken,
			channel:  defTelegramChannel,
			helptext: defTelegramHelptext,
		},
		mastodon: &mastodon{
			server:       defMastodonServer,
			clientKey:    defMastodonClientKey,
			clientSecret: defMastodonClientSecret,
			accessToken:  defMastodonAccessToken,
		},
		discord: &discord{
			appID:    defDiscordBotToken,
			botToken: defDiscordBotToken,
			channel:  defDiscordChannel,
			helptext: defDiscordHelptext,
		},
		twitter: &twitter{
			consumerKey:    defTwitterConsumerKey,
			consumerSecret: defTwitterConsumerSecret,
			accessToken:    defTwitterAccessToken,
			accessSecret:   defTwitterAccessSecret,
		},
		matrix: &matrix{
			homeserver: defMatrixHomeserver,
			userID:     defMatrixUserID,
			roomID:     defMatrixRoomID,
			password:   defMatrixPassword,
		},
		slack: &slack{
			appToken: defSlackAppToken,
			botToken: defSlackBotToken,
			channel:  defSlackChannel,
			helptext: defSlackHelptext,
		},
		github: &github{
			token: defGitHubToken,
			owner: defGitHubOwner,
			repo:  defGitHubRepo,
		},
		notion: &notion{
			token:      defNotionToken,
			databaseID: defNotionDatabaseID,
		},
		nostr: &nostr{
			url:        defNostrRelayURL,
			privateKey: defNostrPrivateKey,
		},
		irc: &irc{
			nick:     defIRCNick,
			password: defIRCPassword,
			channel:  defIRCChannel,
			server:   defIRCServer,
		},
		onion: &onion{
			localPort:   defOnionLocalPort,
			remotePorts: defOnionRemotePorts,
		},
		xmpp: &xmpp{
			username: defXMPPUsername,
			password: defXMPPPassword,
			noTLS:    defXMPPNoTLS,
			helptext: defXMPPHelptext,
		},
	}

	return opts
}

func (o *Options) EnableServices(s ...string) {
	for _, srv := range s {
		switch strings.ToLower(srv) {
		case ServiceDiscord.String():
			o.services.Store(ServiceDiscord, true)
		case ServiceHTTPd.String(), "web":
			o.services.Store(ServiceHTTPd, true)
		case ServiceMastodon.String(), "mstdn":
			o.services.Store(ServiceMastodon, true)
		case ServiceMatrix.String():
			o.services.Store(ServiceMatrix, true)
		case ServiceIRC.String():
			o.services.Store(ServiceIRC, true)
		case ServiceSlack.String():
			o.services.Store(ServiceSlack, true)
		case ServiceTelegram.String():
			o.services.Store(ServiceTelegram, true)
		case ServiceTwitter.String():
			o.services.Store(ServiceTwitter, true)
		case ServiceXMPP.String():
			o.services.Store(ServiceXMPP, true)
		}
	}
}

func (o *Options) isEnabled(f Flag) bool {
	v, ok := o.services.Load(f)
	if !ok {
		return false
	}
	return v.(bool)
}

// HasDebugMode returns true if debug mode is enabled.
func (o *Options) HasDebugMode() bool {
	return o.debug
}

// LogTime returns if the time should be displayed in log messages.
func (o *Options) LogTime() bool {
	return o.logTime
}

// LogLevel returns if the log level.
func (o *Options) LogLevel() logger.LogLevel {
	switch strings.ToLower(o.logLevel) {
	case "info":
		return logger.LevelInfo
	case "warn":
		return logger.LevelWarn
	case "error":
		return logger.LevelError
	case "fatal":
		return logger.LevelFatal
	case "debug":
		return logger.LevelDebug
	default:
		return logger.LevelInfo
	}
}

// EnabledMetrics returns true if metrics collector is enabled.
func (o *Options) EnabledMetrics() bool {
	return o.metrics
}

// IPFSHost returns the host of IPFS daemon service.
func (o *Options) IPFSHost() string {
	return o.ipfs.host
}

// IPFSPort returns the port of IPFS daemon service.
func (o *Options) IPFSPort() int {
	return o.ipfs.port
}

// IPFSMode returns the mode to using IPFS.
func (o *Options) IPFSMode() string {
	return o.ipfs.mode
}

// IPFSTarget returns which IPFS pinning service to use.
func (o *Options) IPFSTarget() string {
	if IPFSToken != "" {
		return IPFSTarget
	}
	return o.ipfs.target
}

// IPFSApiKey returns the apikey of the IPFS pinning service.
// It returns a managed IPFS credential if env `WAYBACK_IPFS_APIKEY` empty.
func (o *Options) IPFSApikey() string {
	if o.ipfs.apikey == "" {
		return IPFSToken
	}
	return o.ipfs.apikey
}

// IPFSSecret returns the secret of the IPFS pinning service.
func (o *Options) IPFSSecret() string {
	return o.ipfs.secret
}

// UseTor returns whether to use the Tor proxy when snapshot webpage.
func (o *Options) UseTor() bool {
	return o.overTor
}

// Slots returns configurations of wayback service, e.g. Internet Archive
func (o *Options) Slots() map[string]bool {
	return o.slots
}

// TelegramToken returns the token of Telegram Bot.
func (o *Options) TelegramToken() string {
	return o.telegram.token
}

// TelegramChannel returns the Telegram Channel name.
func (o *Options) TelegramChannel() string {
	if len(o.telegram.channel) == 0 {
		return ""
	}

	switch o.telegram.channel[:1] {
	case "-", "@":
		return o.telegram.channel
	default:
		return "@" + o.telegram.channel
	}
}

// TelegramHelptext returns the help text for Telegram bot.
func (o *Options) TelegramHelptext() string {
	return breakLine(o.telegram.helptext)
}

// PublishToChannel returns whether to publish results to Telegram Channel.
func (o *Options) PublishToChannel() bool {
	return o.telegram.token != "" && o.telegram.channel != ""
}

// TelegramEnabled returns whether enable Telegram service.
func (o *Options) TelegramEnabled() bool {
	return o.telegram.token != "" && o.isEnabled(ServiceTelegram)
}

// MastodonServer returns the domain of Mastodon instance.
func (o *Options) MastodonServer() string {
	if strings.HasPrefix(o.mastodon.server, "http://") || strings.HasPrefix(o.mastodon.server, "https://") {
		return o.mastodon.server
	}
	o.mastodon.server = "http://" + o.mastodon.server
	u, err := url.Parse(o.mastodon.server)
	if err != nil {
		return ""
	}

	return u.String()
}

// MastodonClientKey returns the client key of Mastodon application.
func (o *Options) MastodonClientKey() string {
	return o.mastodon.clientKey
}

// MastodonClientSecret returns the cilent secret of Mastodon application.
func (o *Options) MastodonClientSecret() string {
	return o.mastodon.clientSecret
}

// MastodonAccessToken returns the access token of Mastodon application.
func (o *Options) MastodonAccessToken() string {
	return o.mastodon.accessToken
}

// PublishToMastodon returns whether to publish result to Mastodon.
func (o *Options) PublishToMastodon() bool {
	return o.MastodonServer() != "" &&
		o.MastodonClientKey() != "" &&
		o.MastodonAccessToken() != "" &&
		o.MastodonClientSecret() != ""
}

// MastodonEnabled returns whether enable Mastodon service.
func (o *Options) MastodonEnabled() bool {
	return o.PublishToMastodon() && o.isEnabled(ServiceMastodon)
}

// TwitterConsumerKey returns the consumer key of Twitter application.
func (o *Options) TwitterConsumerKey() string {
	return o.twitter.consumerKey
}

// TwitterConsumerSecret returns the consumer secret of Twitter application.
func (o *Options) TwitterConsumerSecret() string {
	return o.twitter.consumerSecret
}

// TwitterAccessToken returns the access token of Twitter application.
func (o *Options) TwitterAccessToken() string {
	return o.twitter.accessToken
}

// TwitterAccessSecret returns the access secret of Twitter application.
func (o *Options) TwitterAccessSecret() string {
	return o.twitter.accessSecret
}

// PublishToTwitter returns whether to publish result to Twitter.
func (o *Options) PublishToTwitter() bool {
	return o.TwitterConsumerKey() != "" &&
		o.TwitterConsumerSecret() != "" &&
		o.TwitterAccessToken() != "" &&
		o.TwitterAccessSecret() != ""
}

// TwitterEnabled returns whether enable Twitter service.
func (o *Options) TwitterEnabled() bool {
	return o.PublishToTwitter() && o.isEnabled(ServiceTwitter)
}

// GitHubToken returns the personal access token of GitHub account.
func (o *Options) GitHubToken() string {
	return o.github.token
}

// GitHubOwner returns the user id of GitHub account.
func (o *Options) GitHubOwner() string {
	return o.github.owner
}

// GitHubRepo returns the GitHub repository which to publish results.
func (o *Options) GitHubRepo() string {
	return o.github.repo
}

// PublishToIssues returns whether to publish results to GitHub issues.
func (o *Options) PublishToIssues() bool {
	return o.github.token != "" && o.github.owner != "" && o.github.repo != ""
}

// IRCNick returns nick of IRC
func (o *Options) IRCNick() string {
	return o.irc.nick
}

// IRCPassword returns password of IRC
func (o *Options) IRCPassword() string {
	return o.irc.password
}

// IRCChannel returns channel of IRC
func (o *Options) IRCChannel() string {
	if o.irc.channel == "" {
		return ""
	}
	if strings.HasPrefix(o.irc.channel, "#") {
		return o.irc.channel
	}
	return "#" + o.irc.channel
}

// IRCServer returns server of IRC
func (o *Options) IRCServer() string {
	return o.irc.server
}

// PublishToIRCChannel returns whether publish results to IRC channel.
func (o *Options) PublishToIRCChannel() bool {
	return o.irc.nick != "" && o.irc.channel != ""
}

// IRCEnabled returns whether enable IRC service.
func (o *Options) IRCEnabled() bool {
	return o.irc.nick != "" && o.isEnabled(ServiceIRC)
}

// MatrixHomeserver returns the homeserver of Matrix.
func (o *Options) MatrixHomeserver() string {
	u, err := url.Parse(o.matrix.homeserver)
	if err != nil {
		return ""
	}
	return u.String()
}

// MatrixUserID returns the user ID of Matrix account.
func (o *Options) MatrixUserID() string {
	if !strings.HasPrefix(o.matrix.userID, "@") || !strings.Contains(o.matrix.userID, ":") {
		return ""
	}
	return o.matrix.userID
}

// MatrixRoomID returns the room ID of Matrix account.
func (o *Options) MatrixRoomID() string {
	if !strings.HasPrefix(o.matrix.roomID, "!") || !strings.Contains(o.matrix.roomID, ":") {
		return ""
	}
	return o.matrix.roomID
}

// MatrixPassword returns the password of Matrix account.
func (o *Options) MatrixPassword() string {
	return o.matrix.password
}

// PublishToMatrixRoom returns whether publish results to Matrix room.
func (o *Options) PublishToMatrixRoom() bool {
	return o.MatrixHomeserver() != "" &&
		o.MatrixUserID() != "" &&
		o.MatrixRoomID() != "" &&
		o.MatrixPassword() != ""
}

// MatrixEnabled returns whether enable Matrix service.
func (o *Options) MatrixEnabled() bool {
	return o.MatrixHomeserver() != "" &&
		o.MatrixUserID() != "" &&
		o.MatrixPassword() != "" &&
		o.isEnabled(ServiceMatrix)
}

// DiscordBotToken returns the token of Discord bot
func (o *Options) DiscordBotToken() string {
	return o.discord.botToken
}

// DiscordChannel returns the channel on Discord.
func (o *Options) DiscordChannel() string {
	// if strings.HasPrefix(o.discord.channel, "#") {
	// 	return o.discord.channel
	// }
	// if o.discord.channel != "" {
	// 	return "#" + o.discord.channel
	// }
	return o.discord.channel
}

// DiscordHelptext returns the help text for Discord bot
func (o *Options) DiscordHelptext() string {
	return breakLine(o.discord.helptext)
}

// PublishToDiscordChannel returns whether publish results to Discord channel.
func (o *Options) PublishToDiscordChannel() bool {
	return o.DiscordBotToken() != "" && o.DiscordChannel() != ""
}

// DiscordEnabled returns whether enable Discord service.
func (o *Options) DiscordEnabled() bool {
	return o.DiscordBotToken() != "" && o.isEnabled(ServiceDiscord)
}

// SlackAppToken returns the app-level token of Slack bot.
func (o *Options) SlackAppToken() string {
	return o.slack.appToken
}

// SlackBotToken returns the bot user auth token of Slack bot.
func (o *Options) SlackBotToken() string {
	return o.slack.botToken
}

// SlackChannel returns the Slack channel id.
func (o *Options) SlackChannel() string {
	return o.slack.channel
}

// SlackHelptext returns the help text for Slack bot.
func (o *Options) SlackHelptext() string {
	return breakLine(o.slack.helptext)
}

// PublishToSlackChannel returns whether publish results to Slack channel.
func (o *Options) PublishToSlackChannel() bool {
	return o.SlackBotToken() != "" && o.SlackChannel() != ""
}

// SlackEnabled returns whether enable Slack service.
func (o *Options) SlackEnabled() bool {
	return o.SlackBotToken() != "" && o.isEnabled(ServiceSlack)
}

// XMPPUsername returns the XMPP username (JID).
func (o *Options) XMPPUsername() string {
	return o.xmpp.username
}

// XMPPPassword returns the XMPP password.
func (o *Options) XMPPPassword() string {
	return o.xmpp.password
}

// XMPPNoTLS returns whether disable TLS.
func (o *Options) XMPPNoTLS() bool {
	return o.xmpp.noTLS
}

// XMPPHelptext returns the help text for XMPP.
func (o *Options) XMPPHelptext() string {
	return breakLine(o.xmpp.helptext)
}

// XMPPEnabled returns whether enable XMPP service.
func (o *Options) XMPPEnabled() bool {
	return o.XMPPUsername() != "" && o.XMPPPassword() != "" && o.isEnabled(ServiceXMPP)
}

// NotionToken returns the Notion integration token.
func (o *Options) NotionToken() string {
	return o.notion.token
}

// NotionDatabaseID returns the Notion's dabase id.
func (o *Options) NotionDatabaseID() string {
	return o.notion.databaseID
}

// PublishToNotion determines whether the results should be published on Notion.
func (o *Options) PublishToNotion() bool {
	return o.NotionToken() != "" && o.NotionDatabaseID() != ""
}

// NostrRelayURL returns the relay url of Nostr server.
func (o *Options) NostrRelayURL() []string {
	return strings.Split(o.nostr.url, ",")
}

// NostrPrivateKey returns the private key of Nostr account.
func (o *Options) NostrPrivateKey() string {
	return o.nostr.privateKey
}

// PublishToNostr determines whether the results should be published on Nostr.
func (o *Options) PublishToNostr() bool {
	return len(o.NostrRelayURL()) > 0 && o.NostrPrivateKey() != ""
}

// OnionPrivKey returns the private key of Onion service.
func (o *Options) OnionPrivKey() string {
	return o.onion.pvk
}

// OnionLocalPort returns the local port to a TCP listener on.
// This is ignored if `WAYBACK_LISTEN_ADDR` is set.
func (o *Options) OnionLocalPort() int {
	return o.onion.localPort
}

// OnionRemotePorts returns the remote ports to serve the Onion hidden service on.
func (o *Options) OnionRemotePorts() []int {
	return o.onion.remotePorts
}

// OnionDisabled returns whether disable Onion service.
func (o *Options) OnionDisabled() bool {
	return o.onion.disabled
}

// ListenAddr returns the listen address for the HTTP server.
func (o *Options) ListenAddr() string {
	return o.listenAddr
}

// EnabledChromeRemote returns whether enable Chrome/Chromium remote debugging
// for screenshot
func (o *Options) EnabledChromeRemote() bool {
	return o.enabledChromeRemote
}

// ChromeRemoteAddr returns the remote debugging address for Chrome/Chromium
func (o *Options) ChromeRemoteAddr() string {
	return o.chromeRemoteAddr
}

// BoltPathname returns filename of bolt database
func (o *Options) BoltPathname() string {
	return o.boltPathname
}

// PoolingSize returns the number of worker pool
func (o *Options) PoolingSize() int {
	return o.poolingSize
}

// StorageDir returns the directory to storage binary file, e.g. html file, PDF
func (o *Options) StorageDir() string {
	return o.storageDir
}

// EnabledReduxer returns whether enable store binary file locally.
func (o *Options) EnabledReduxer() bool {
	return o.StorageDir() != ""
}

// MaxMediaSize returns max size to limit download stream media.
func (o *Options) MaxMediaSize() uint64 {
	size, err := humanize.ParseBytes(o.maxMediaSize)
	if err != nil {
		return 0
	}
	return size
}

// MaxAttachSize returns max attach size limits for several services.
// scope: telegram
func (o *Options) MaxAttachSize(scope string) int64 {
	scopes := map[string]int64{
		"telegram": maxAttachSizeTelegram,
		"discord":  maxAttachSizeDiscord,
		"slack":    maxAttachSizeSlack,
	}
	return scopes[scope]
}

// WaybackTimeout returns timeout for a wayback request.
func (o *Options) WaybackTimeout() time.Duration {
	return time.Duration(o.waybackTimeout) * time.Second
}

// WaybackMaxRetries returns max retries for a wayback request.
func (o *Options) WaybackMaxRetries() uint64 {
	return uint64(o.waybackMaxRetries)
}

// WaybackUserAgent returns User-Agent for a wayback request.
func (o *Options) WaybackUserAgent() string {
	return o.waybackUserAgent
}

// WaybackFallback returns whether fallback to Google cache is enabled if
// the original webpage is unavailable.
func (o *Options) WaybackFallback() bool {
	return o.waybackFallback
}

// WaybackMeiliEndpoint returns the Meilisearch API endpoint.
func (o *Options) WaybackMeiliEndpoint() string {
	return o.waybackMeiliEndpoint
}

// WaybackMeiliIndexing returns the Meilisearch indexing name.
func (o *Options) WaybackMeiliIndexing() string {
	return o.waybackMeiliIndexing
}

// WaybackMeiliApikey returns the Meilisearch admin apikey.
func (o *Options) WaybackMeiliApikey() string {
	return o.waybackMeiliApikey
}

// EnabledMeilisearch returns whether enable meilisearch server.
func (o *Options) EnabledMeilisearch() bool {
	return o.WaybackMeiliEndpoint() != ""
}

// HTTPdEnabled returns whether enable HTTP daemon service.
func (o *Options) HTTPdEnabled() bool {
	return o.isEnabled(ServiceHTTPd)
}

// Proxy returns the proxy server address.
func (o *Options) Proxy() string {
	return o.proxy
}
