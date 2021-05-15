// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package config // import "github.com/wabarc/wayback/config"

import (
	"net/url"
	"strings"
)

const (
	defDebug    = false
	defLogTime  = true
	defMetrics  = false
	defOverTor  = false
	defIPFSHost = "127.0.0.1"
	defIPFSPort = 4001
	defIPFSMode = "pinner"

	defEnabledIA = false
	defEnabledIS = false
	defEnabledIP = false
	defEnabledPH = false

	defTelegramToken    = ""
	defTelegramChannel  = ""
	defTelegramHelptext = ""
	defGitHubToken      = ""
	defGitHubOwner      = ""
	defGitHubRepo       = ""

	defMastodonServer        = "https://mastodon.social"
	defMastodonClientKey     = ""
	defMastodonClientSecret  = ""
	defMastodonAccessToken   = ""
	defTwitterConsumerKey    = ""
	defTwitterConsumerSecret = ""
	defTwitterAccessToken    = ""
	defTwitterAccessSecret   = ""

	defIRCNick     = ""
	defIRCPassword = ""
	defIRCChannel  = ""
	defIRCServer   = "irc.freenode.net:7000"

	defMatrixHomeserver = "https://matrix.org"
	defMatrixUserID     = ""
	defMatrixRoomID     = ""
	defMatrixPassword   = ""

	defTorPrivateKey = ""
	defTorLocalPort  = 8964
	defTorrcFile     = "/etc/tor/torrc"

	defChromeRemoteAddr    = ""
	defEnabledChromeRemote = false
	defBoltFilename        = "wayback.db"
	defPoolingSize         = 3
)

var (
	defTorRemotePorts = []int{80}
)

type Options struct {
	debug   bool
	logTime bool
	overTor bool
	metrics bool

	ipfs     *ipfs
	slots    map[string]bool
	telegram *telegram
	mastodon *mastodon
	twitter  *twitter
	github   *github
	matrix   *matrix
	irc      *irc
	tor      *tor

	chromeRemoteAddr    string
	enabledChromeRemote bool
	poolingSize         int
}

type ipfs struct {
	host string
	port uint
	mode string
}

type slots struct {
	ia bool
	is bool
	ip bool
	ph bool
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

type matrix struct {
	homeserver string
	userID     string
	roomID     string
	password   string
}

type irc struct {
	nick     string
	password string
	channel  string
	server   string
}

type tor struct {
	pvk string

	localPort   int
	remotePorts []int
	torrcFile   string
}

// NewOptions returns Options with default values.
func NewOptions() *Options {
	opts := &Options{
		debug:               defDebug,
		logTime:             defLogTime,
		overTor:             defOverTor,
		metrics:             defMetrics,
		chromeRemoteAddr:    defChromeRemoteAddr,
		enabledChromeRemote: defEnabledChromeRemote,
		poolingSize:         defPoolingSize,
		ipfs: &ipfs{
			host: defIPFSHost,
			port: defIPFSPort,
			mode: defIPFSMode,
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
		github: &github{
			token: defGitHubToken,
			owner: defGitHubOwner,
			repo:  defGitHubRepo,
		},
		irc: &irc{
			nick:     defIRCNick,
			password: defIRCPassword,
			channel:  defIRCChannel,
			server:   defIRCServer,
		},
		tor: &tor{
			pvk:         defTorPrivateKey,
			localPort:   defTorLocalPort,
			remotePorts: defTorRemotePorts,
			torrcFile:   defTorrcFile,
		},
	}

	return opts
}

// HasDebugMode returns true if debug mode is enabled.
func (o *Options) HasDebugMode() bool {
	return o.debug
}

// LogTime returns if the time should be displayed in log messages.
func (o *Options) LogTime() bool {
	return o.logTime
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
func (o *Options) IPFSPort() uint {
	return o.ipfs.port
}

// IPFSMode returns the mode to using IPFS.
func (o *Options) IPFSMode() string {
	return o.ipfs.mode
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
	return o.telegram.channel
}

// TelegramHelptext returns the help text for Telegram bot.
func (o *Options) TelegramHelptext() string {
	o.telegram.helptext = strings.Replace(o.telegram.helptext, `\r`, "\n", -1)
	o.telegram.helptext = strings.Replace(o.telegram.helptext, `\n`, "\n", -1)
	o.telegram.helptext = strings.Replace(o.telegram.helptext, `\r\n`, "\n", -1)
	o.telegram.helptext = strings.Replace(o.telegram.helptext, `<br>`, "\n", -1)
	o.telegram.helptext = strings.Replace(o.telegram.helptext, `<br/>`, "\n", -1)
	return o.telegram.helptext
}

// PublishToChannel returns whether to publish results to Telegram Channel.
func (o *Options) PublishToChannel() bool {
	return o.telegram.token != "" && o.telegram.channel != ""
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

// TorPrivKey returns the private key of Tor service.
func (o *Options) TorPrivKey() string {
	return o.tor.pvk
}

// TorLocalPort returns the local port to a TCP listener on.
func (o *Options) TorLocalPort() int {
	return o.tor.localPort
}

// TorRemotePorts returns the remote ports to serve the Tor hidden service on.
func (o *Options) TorRemotePorts() []int {
	return o.tor.remotePorts
}

// TorrcFile returns path of the torrc file to set on start Tor Hidden Service.
func (o *Options) TorrcFile() string {
	return o.tor.torrcFile
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

// BoltFilename returns filename of bolt database
func (o *Options) BoltFilename() string {
	return defBoltFilename
}

// PoolingSize returns the number of worker pool
func (o *Options) PoolingSize() int {
	return o.poolingSize
}
