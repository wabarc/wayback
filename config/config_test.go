// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

/*
Package config handles configuration management for the application.
*/

package config // import "github.com/wabarc/wayback/config"

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/wabarc/logger"
)

func TestSlotName(t *testing.T) {
	expected := "Internet Archive"
	got := SlotName(SLOT_IA)

	if got != expected {
		t.Fatalf(`Unexpected get the slot name description, got %v instead of %s`, got, expected)
	}
}

func TestSlotNameNotExist(t *testing.T) {
	expected := UNKNOWN
	got := SlotName("something")

	if got != expected {
		t.Fatalf(`Unexpected get the slot name description, got %v instead of %s`, got, expected)
	}
}

func TestAutoSetEnv(t *testing.T) {
	key := "DO_NOT_EXIST"
	val := "yes"
	os.Clearenv()

	tmpfile, err := os.CreateTemp("", "wayback-cfg-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := []byte(key + "=" + val)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	if _, err := NewParser().ParseFile(tmpfile.Name()); err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := os.Getenv(key)
	if got != val {
		t.Fatalf(`Unexpected set environment, got %v instead of %v`, got, val)
	}
}

func TestDebugModeOn(t *testing.T) {
	os.Clearenv()
	os.Setenv("DEBUG", "on")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := true
	got := opts.HasDebugMode()

	if got != expected {
		t.Fatalf(`Unexpected debug mode value, got %v instead of %v`, got, expected)
	}
}

func TestDebugModeOff(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := false
	got := opts.HasDebugMode()

	if got != expected {
		t.Fatalf(`Unexpected debug mode value, got %v instead of %v`, got, expected)
	}
}

func TestEnableLogTime(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := true
	got := opts.LogTime()

	if got != expected {
		t.Fatalf(`Unexpected logging time, got %v instead of %v`, got, expected)
	}
}

func TestDisableLogTime(t *testing.T) {
	os.Clearenv()
	os.Setenv("LOG_TIME", "false")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := false
	got := opts.LogTime()

	if got != expected {
		t.Fatalf(`Unexpected logging time, got %v instead of %v`, got, expected)
	}
}

func TestMetricsEnabled(t *testing.T) {
	os.Clearenv()
	os.Setenv("ENABLE_METRICS", "on")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := true
	got := opts.EnabledMetrics()

	if got != expected {
		t.Fatalf(`Unexpected metrics mode value, got %v instead of %v`, got, expected)
	}
}

func TestMetricsDisabled(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := false
	got := opts.EnabledMetrics()

	if got != expected {
		t.Fatalf(`Unexpected metrics mode value, got %v instead of %v`, got, expected)
	}
}

func TestIPFSHost(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_IPFS_HOST", "127.0.0.1")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "127.0.0.1"
	got := opts.IPFSHost()

	if got != expected {
		t.Fatalf(`Unexpected IPFS host, got %v instead of %s`, got, expected)
	}
}

func TestIPFSPort(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_IPFS_PORT", "1234")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := 1234
	got := opts.IPFSPort()

	if got != expected {
		t.Fatalf(`Unexpected IPFS port, got %v instead of %q`, got, expected)
	}
}

func TestIPFSMode(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_IPFS_MODE", "mode")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "mode"
	got := opts.IPFSMode()

	if got != expected {
		t.Fatalf(`Unexpected IPFS mode, got %v instead of %s`, got, expected)
	}
}

func TestIPFSTarget(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_IPFS_TARGET", "target")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "target"
	got := opts.IPFSTarget()

	if got != expected {
		t.Fatalf(`Unexpected IPFS target, got %v instead of %s`, got, expected)
	}
}

func TestIPFSApikey(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_IPFS_APIKEY", "apikey")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "apikey"
	got := opts.IPFSApikey()

	if got != expected {
		t.Fatalf(`Unexpected IPFS apikey, got %v instead of %s`, got, expected)
	}
}

func TestIPFSSecret(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_IPFS_SECRET", "secret")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "secret"
	got := opts.IPFSSecret()

	if got != expected {
		t.Fatalf(`Unexpected IPFS secret, got %v instead of %s`, got, expected)
	}
}

func TestOverTor(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_USE_TOR", "true")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := true
	got := opts.UseTor()

	if got != expected {
		t.Fatalf(`Unexpected over Tor, got %v instead of %v`, got, expected)
	}
}

func TestEnableSlots(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_ENABLE_IA", "true")
	os.Setenv("WAYBACK_ENABLE_IS", "true")
	os.Setenv("WAYBACK_ENABLE_IP", "true")
	os.Setenv("WAYBACK_ENABLE_PH", "true")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := map[string]bool{
		SLOT_IA: true,
		SLOT_IS: true,
		SLOT_IP: true,
		SLOT_PH: true,
	}
	got := opts.Slots()

	if got == nil || !got[SLOT_IA] || !got[SLOT_IS] || !got[SLOT_IP] || !got[SLOT_PH] {
		t.Fatalf(`Unexpected over Tor, got %v instead of %v`, got, expected)
	}
}

func TestDefaultSlots(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := map[string]bool{
		SLOT_IA: false,
		SLOT_IS: false,
		SLOT_IP: false,
	}
	got := opts.Slots()

	if got == nil || got[SLOT_IA] || got[SLOT_IS] || got[SLOT_IP] {
		t.Fatalf(`Unexpected over Tor, got %v instead of %v`, got, expected)
	}
}

func TestTelegramToken(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_TELEGRAM_TOKEN", "tg:token")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "tg:token"
	got := opts.TelegramToken()

	if got != expected {
		t.Fatalf(`Unexpected Telegram Bot token, got %v instead of %s`, got, expected)
	}
}

func TestTelegramChannel(t *testing.T) {
	var tests = []struct {
		name string
		expt string
	}{
		{
			name: "",
			expt: "",
		},
		{
			name: "tgchannelname",
			expt: "@tgchannelname",
		},
		{
			name: "@tgchannelname",
			expt: "@tgchannelname",
		},
		{
			name: "-123456",
			expt: "-123456",
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_TELEGRAM_CHANNEL", test.name)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			expected := test.expt
			got := opts.TelegramChannel()

			if got != expected {
				t.Fatalf(`Unexpected Telegram channel name, got %v instead of %s`, got, expected)
			}
		})
	}
}

func TestTelegramHelptext(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_TELEGRAM_HELPTEXT", "some text")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "some text"
	got := opts.TelegramHelptext()

	if got != expected {
		t.Fatalf(`Unexpected Telegram help text, got %v instead of %s`, got, expected)
	}
}

func TestTorPrivateKey(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_TOR_PRIVKEY", "tor:private:key")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "tor:private:key"
	got := opts.TorPrivKey()

	if got != expected {
		t.Fatalf(`Unexpected Tor private key, got %v instead of %s`, got, expected)
	}
}

func TestTorLocalPort(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_TOR_LOCAL_PORT", "8080")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := 8080
	got := opts.TorLocalPort()

	if got != expected {
		t.Fatalf(`Unexpected Tor local port, got %v instead of %q`, got, expected)
	}
}

func TestDefaultTorLocalPortValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := defTorLocalPort
	got := opts.TorLocalPort()

	if got != expected {
		t.Fatalf(`Unexpected Tor local port, got %v instead of %q`, got, expected)
	}
}

func TestTorRemotePorts(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_TOR_REMOTE_PORTS", "80,81,82")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := []int{80, 81, 82}
	got := opts.TorRemotePorts()

	if got == nil || len(got) != 3 {
		t.Fatalf(`Unexpected Tor remote port, got %v instead of %v`, got, expected)
	}
}

func TestListenAddr(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		address  string
		expected string
	}{
		{
			address:  "",
			expected: defListenAddr,
		},
		{
			address:  defListenAddr,
			expected: defListenAddr,
		},
	}

	for _, test := range tests {
		t.Run(test.address, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("HTTP_LISTEN_ADDR", test.address)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing failure: %v`, err)
			}

			result := opts.ListenAddr()
			if result != test.expected {
				t.Fatalf(`Unexpected LISTEN_ADDR value, got %q instead of %q`, result, test.expected)
			}
		})
	}
}

func TestDefaultTorRemotePortsValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := []int{80}
	got := opts.TorRemotePorts()

	if got == nil || len(got) != 1 {
		t.Fatalf(`Unexpected Tor remote port, got %v instead of %v`, got, expected)
	}
}

func TestGitHubToken(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_GITHUB_TOKEN", "github:token")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "github:token"
	got := opts.GitHubToken()

	if got != expected {
		t.Fatalf(`Unexpected GitHub personal access token, got %v instead of %s`, got, expected)
	}
}

func TestGitHubOwner(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_GITHUB_OWNER", "github-owner")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "github-owner"
	got := opts.GitHubOwner()

	if got != expected {
		t.Fatalf(`Unexpected GitHub owner, got %v instead of %s`, got, expected)
	}
}

func TestGitHubRepo(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_GITHUB_REPO", "github-repo")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "github-repo"
	got := opts.GitHubRepo()

	if got != expected {
		t.Fatalf(`Unexpected GitHub repository, got %v instead of %s`, got, expected)
	}
}

func TestNotionToken(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_NOTION_TOKEN", "notion:token")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "notion:token"
	got := opts.NotionToken()

	if got != expected {
		t.Fatalf(`Unexpected Notion integration token, got %v instead of %s`, got, expected)
	}
}

func TestNotionDatabaseID(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_NOTION_DATABASE_ID", "uuid4")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "uuid4"
	got := opts.NotionDatabaseID()

	if got != expected {
		t.Fatalf(`Unexpected Notion's database id, got %v instead of %s`, got, expected)
	}
}

func TestMastodonServer(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_MASTODON_SERVER", "https://mastodon.social")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "https://mastodon.social"
	got := opts.MastodonServer()

	if got != expected {
		t.Fatalf(`Unexpected Mastodon instance domain, got %v instead of %s`, got, expected)
	}
}

func TestMastodonClientKey(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_MASTODON_KEY", "foo")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "foo"
	got := opts.MastodonClientKey()

	if got != expected {
		t.Fatalf(`Unexpected Mastodon client key, got %v instead of %s`, got, expected)
	}
}

func TestMastodonClientSecret(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_MASTODON_SECRET", "foo")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "foo"
	got := opts.MastodonClientSecret()

	if got != expected {
		t.Fatalf(`Unexpected Mastodon client secret, got %v instead of %s`, got, expected)
	}
}

func TestMastodonAccessToken(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_MASTODON_TOKEN", "foo")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "foo"
	got := opts.MastodonAccessToken()

	if got != expected {
		t.Fatalf(`Unexpected Mastodon access token, got %v instead of %s`, got, expected)
	}
}

func TestIRCNick(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_IRC_NICK", "foo")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "foo"
	got := opts.IRCNick()

	if got != expected {
		t.Fatalf(`Unexpected IRC nick got %v instead of %s`, got, expected)
	}
}

func TestIRCPassword(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_IRC_PASSWORD", "foo")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "foo"
	got := opts.IRCPassword()

	if got != expected {
		t.Fatalf(`Unexpected IRC password got %v instead of %s`, got, expected)
	}
}

func TestIRCChannel(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_IRC_CHANNEL", "foo")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "#foo"
	got := opts.IRCChannel()

	if got != expected {
		t.Fatalf(`Unexpected IRC channel got %v instead of %s`, got, expected)
	}
}

func TestIRCServer(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_IRC_SERVER", "example.net:7000")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "example.net:7000"
	got := opts.IRCServer()

	if got != expected {
		t.Fatalf(`Unexpected IRC server got %v instead of %s`, got, expected)
	}
}

func TestPublishToIRCChannel(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_IRC_NICK", "foo")
	os.Setenv("WAYBACK_IRC_CHANNEL", "bar")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := true
	got := opts.PublishToIRCChannel()

	if got != expected {
		t.Fatalf(`Unexpected publish to IRC channel got %t instead of %v`, got, expected)
	}
}

func TestMatrixHomeServer(t *testing.T) {
	expected := "https://matrix-client.matrix.org"

	os.Clearenv()
	os.Setenv("WAYBACK_MATRIX_HOMESERVER", expected)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.MatrixHomeserver()
	if got != expected {
		t.Fatalf(`Unexpected Matrix homeserver got %v instead of %v`, got, expected)
	}
}

func TestMatrixUserID(t *testing.T) {
	expected := "@foo:matrix.org"

	os.Clearenv()
	os.Setenv("WAYBACK_MATRIX_USERID", expected)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.MatrixUserID()
	if got != expected {
		t.Fatalf(`Unexpected Matrix user ID got %v instead of %v`, got, expected)
	}
}

func TestMatrixRoomID(t *testing.T) {
	expected := "!foo:matrix.org"

	os.Clearenv()
	os.Setenv("WAYBACK_MATRIX_ROOMID", expected)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.MatrixRoomID()
	if got != expected {
		t.Fatalf(`Unexpected Matrix room ID got %v instead of %v`, got, expected)
	}
}

func TestMatrixPassword(t *testing.T) {
	expected := "foo-bar"

	os.Clearenv()
	os.Setenv("WAYBACK_MATRIX_PASSWORD", expected)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.MatrixPassword()
	if got != expected {
		t.Fatalf(`Unexpected Matrix password got %v instead of %v`, got, expected)
	}
}

func TestPublishToMatrixRoom(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_MATRIX_HOMESERVER", "https://matrix-client.matrix.org")
	os.Setenv("WAYBACK_MATRIX_USERID", "@foo:matrix.org")
	os.Setenv("WAYBACK_MATRIX_ROOMID", "!bar:matrix.org")
	os.Setenv("WAYBACK_MATRIX_PASSWORD", "zoo")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := true
	got := opts.PublishToMatrixRoom()

	if got != expected {
		t.Fatalf(`Unexpected publish to Matrix room got %t instead of %v`, got, expected)
	}
}

func TestDiscordBotToken(t *testing.T) {
	expected := "foo-bar"

	os.Clearenv()
	os.Setenv("WAYBACK_DISCORD_BOT_TOKEN", expected)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.DiscordBotToken()
	if got != expected {
		t.Fatalf(`Unexpected Discord bot token got %v instead of %v`, got, expected)
	}
}

func TestDiscordHelptext(t *testing.T) {
	expected := "some text"

	os.Clearenv()
	os.Setenv("WAYBACK_DISCORD_HELPTEXT", expected)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.DiscordHelptext()
	if got != expected {
		t.Fatalf(`Unexpected Discord help text got %v instead of %v`, got, expected)
	}
}

func TestDiscordChannel(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		channel, expected string
	}{
		{
			channel:  "",
			expected: "",
		},
		{
			channel:  "865981235815140000",
			expected: "865981235815140000",
		},
	}

	for _, test := range tests {
		t.Run(test.channel, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_DISCORD_CHANNEL", test.channel)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.DiscordChannel()
			if got != test.expected {
				t.Fatalf(`Unexpected Discord channel got %v instead of %v`, got, test.expected)
			}
		})
	}
}

func TestPublishToDiscordChannel(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_DISCORD_BOT_TOKEN", "discord-bot-token")
	os.Setenv("WAYBACK_DISCORD_CHANNEL", "865981235815140000")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := true
	got := opts.PublishToDiscordChannel()

	if got != expected {
		t.Fatalf(`Unexpected publish to Discord channel got %t instead of %v`, got, expected)
	}
}

func TestSlackAppToken(t *testing.T) {
	expected := "xapp-1-A0000000FC7-2300600000035-a000000bc7d104f053f66000000e589dafabcde70c5152abaacbcaea00000000"

	os.Clearenv()
	os.Setenv("WAYBACK_SLACK_APP_TOKEN", expected)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.SlackAppToken()
	if got != expected {
		t.Fatalf(`Unexpected Slack app-level token got %v instead of %v`, got, expected)
	}
}

func TestSlackBotToken(t *testing.T) {
	expected := "xoxb-2306408000000-2300127000000-GgLHgzqK3fXH5KA50AAbcdef"

	os.Clearenv()
	os.Setenv("WAYBACK_SLACK_BOT_TOKEN", expected)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.SlackBotToken()
	if got != expected {
		t.Fatalf(`Unexpected Slack bot token got %v instead of %v`, got, expected)
	}
}

func TestSlackChannel(t *testing.T) {
	expected := "C123ABCXY89"

	os.Clearenv()
	os.Setenv("WAYBACK_SLACK_CHANNEL", expected)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.SlackChannel()
	if got != expected {
		t.Fatalf(`Unexpected Slack channel id got %v instead of %v`, got, expected)
	}
}

func TestSlackHelptext(t *testing.T) {
	expected := "some text"

	os.Clearenv()
	os.Setenv("WAYBACK_SLACK_HELPTEXT", expected)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.SlackHelptext()
	if got != expected {
		t.Fatalf(`Unexpected Slack help text got %v instead of %v`, got, expected)
	}
}

func TestPublishToSlackChannel(t *testing.T) {
	os.Clearenv()
	// os.Setenv("WAYBACK_SLACK_APP_TOKEN", "slack-app-token")
	os.Setenv("WAYBACK_SLACK_BOT_TOKEN", "slack-bot-token")
	os.Setenv("WAYBACK_SLACK_CHANNEL", "C123ABCXY89")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := true
	got := opts.PublishToSlackChannel()

	if got != expected {
		t.Fatalf(`Unexpected publish to Slack channel got %t instead of %v`, got, expected)
	}
}

func TestEnabledChromeRemote(t *testing.T) {
	addr := "127.0.0.1:1234"

	os.Clearenv()
	os.Setenv("CHROME_REMOTE_ADDR", addr)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	yes := opts.EnabledChromeRemote()
	if !yes {
		t.Fatalf(`Unexpected enable Chrome remote debugging got %t instead of 'true'`, yes)
	}
}

func TestDisabledChromeRemote(t *testing.T) {
	os.Clearenv()
	os.Setenv("CHROME_REMOTE_ADDR", "")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	yes := opts.EnabledChromeRemote()
	if yes {
		t.Fatalf(`Unexpected enable Chrome remote debugging got %t instead of 'true'`, yes)
	}
}

func TestChromeRemoteAddr(t *testing.T) {
	addr := "127.0.0.1:1234"

	os.Clearenv()
	os.Setenv("CHROME_REMOTE_ADDR", addr)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.ChromeRemoteAddr()
	if got != addr {
		t.Fatalf(`Unexpected Chrome remote debugging address got %s instead of %s`, got, addr)
	}
}

func TestPoolingSize(t *testing.T) {
	size := 10
	os.Clearenv()
	os.Setenv("WAYBACK_POOLING_SIZE", strconv.Itoa(size))

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.PoolingSize()
	if got != size {
		t.Fatalf(`Unexpected pooling size got %d instead of %d`, got, size)
	}
}

func TestBoltPath(t *testing.T) {
	path := "./wayback.db"

	os.Clearenv()
	os.Setenv("WAYBACK_BOLT_PATH", path)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.BoltPathname()
	if got != path {
		t.Fatalf(`Unexpected bolt db file path got %s instead of %s`, got, path)
	}
}

func TestStorageDir(t *testing.T) {
	dir := os.TempDir()

	os.Clearenv()
	os.Setenv("WAYBACK_STORAGE_DIR", dir)

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	got := opts.StorageDir()
	if got != dir {
		t.Fatalf(`Unexpected storage binary directory got %s instead of %s`, got, dir)
	}
}

func TestEnabledReduxer(t *testing.T) {
	var tests = []struct {
		dir string
		exp bool
	}{
		{
			dir: "",
			exp: false,
		},
		{
			dir: "/path/to/storage",
			exp: true,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_STORAGE_DIR", test.dir)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.EnabledReduxer()
			if got != test.exp {
				t.Fatalf(`Unexpected enabled reduxer got %t instead of %t`, got, test.exp)
			}
		})
	}
}

func TestLogLevel(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		level    string
		expected logger.LogLevel
	}{
		{
			level:    "",
			expected: logger.LevelInfo,
		},
		{
			level:    "info",
			expected: logger.LevelInfo,
		},
		{
			level:    "warn",
			expected: logger.LevelWarn,
		},
		{
			level:    "error",
			expected: logger.LevelError,
		},
		{
			level:    "fatal",
			expected: logger.LevelFatal,
		},
		{
			level:    "unknown",
			expected: logger.LevelInfo,
		},
	}

	for _, test := range tests {
		t.Run(test.level, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("LOG_LEVEL", test.level)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.LogLevel()
			if got != test.expected {
				t.Fatalf(`Unexpected set log level got %d instead of %d`, got, test.expected)
			}
		})
	}
}

func TestMaxMediaSize(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		size     string
		expected uint64
	}{
		{
			size:     "",
			expected: 512000000,
		},
		{
			size:     "invalid",
			expected: 0,
		},
		{
			size:     "10KB",
			expected: 10000,
		},
	}

	for _, test := range tests {
		t.Run(test.size, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_MAX_MEDIA_SIZE", test.size)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.MaxMediaSize()
			if got != test.expected {
				t.Fatalf(`Unexpected set max media size got %d instead of %d`, got, test.expected)
			}
		})
	}
}

func TestWaybackTimeout(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		timeout  int
		expected time.Duration
	}{
		{
			timeout:  0,
			expected: 0 * time.Second,
		},
		{
			timeout:  1,
			expected: time.Second,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_TIMEOUT", strconv.Itoa(test.timeout))

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.WaybackTimeout()
			if got != test.expected {
				t.Fatalf(`Unexpected set wayback timeout got %d instead of %d`, got, test.expected)
			}
		})
	}
}

func TestWaybackMaxRetries(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		timeout  int
		expected uint64
	}{
		{
			timeout:  0,
			expected: 0,
		},
		{
			timeout:  1,
			expected: 1,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_MAX_RETRIES", strconv.Itoa(test.timeout))

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.WaybackMaxRetries()
			if got != test.expected {
				t.Fatalf(`Unexpected set max retires for a wayback request got %d instead of %d`, got, test.expected)
			}
		})
	}
}

func TestWaybackUserAgent(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		userAgent string
		expected  string
	}{
		{
			userAgent: "",
			expected:  defWaybackUserAgent,
		},
		{
			userAgent: "foo bar",
			expected:  "foo bar",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_USERAGENT", test.userAgent)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.WaybackUserAgent()
			if got != test.expected {
				t.Fatalf(`Unexpected set wayback user agent got %s instead of %s`, got, test.expected)
			}
		})
	}
}

func TestWaybackFallback(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		fallback string
		expected bool
	}{
		{
			fallback: "",
			expected: defWaybackFallback,
		},
		{
			fallback: "unexpected",
			expected: defWaybackFallback,
		},
		{
			fallback: "on",
			expected: true,
		},
		{
			fallback: "true",
			expected: true,
		},
		{
			fallback: "0",
			expected: false,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_FALLBACK", test.fallback)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.WaybackFallback()
			if got != test.expected {
				t.Fatalf(`Unexpected set wayback fallback got %t instead of %t`, got, test.expected)
			}
		})
	}
}

func TestWaybackMeiliEndpoint(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		endpoint string
		expected string
	}{
		{
			endpoint: "",
			expected: defWaybackMeiliEndpoint,
		},
		{
			endpoint: "https://example.com",
			expected: "https://example.com",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_MEILI_ENDPOINT", test.endpoint)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.WaybackMeiliEndpoint()
			if got != test.expected {
				t.Fatalf(`Unexpected set meilisearch endpoint got %s instead of %s`, got, test.expected)
			}
		})
	}
}

func TestWaybackMeiliIndexing(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		indexing string
		expected string
	}{
		{
			indexing: "",
			expected: defWaybackMeiliIndexing,
		},
		{
			indexing: "foo-bar",
			expected: "foo-bar",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_MEILI_INDEXING", test.indexing)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.WaybackMeiliIndexing()
			if got != test.expected {
				t.Fatalf(`Unexpected set meilisearch indexing got %s instead of %s`, got, test.expected)
			}
		})
	}
}

func TestWaybackMeiliApikey(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		apikey   string
		expected string
	}{
		{
			apikey:   "",
			expected: defWaybackMeiliApikey,
		},
		{
			apikey:   "foo.bar",
			expected: "foo.bar",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_MEILI_APIKEY", test.apikey)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.WaybackMeiliApikey()
			if got != test.expected {
				t.Fatalf(`Unexpected set meilisearch api key got %s instead of %s`, got, test.expected)
			}
		})
	}
}

func TestEnabledMeilisearch(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		endpoint string
		expected bool
	}{
		{
			endpoint: "",
			expected: false,
		},
		{
			endpoint: "https://example.com",
			expected: true,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_MEILI_ENDPOINT", test.endpoint)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.EnabledMeilisearch()
			if got != test.expected {
				t.Fatalf(`Unexpected enabled meilisearch got %t instead of %t`, got, test.expected)
			}
		})
	}
}
