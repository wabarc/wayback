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

	expected := uint(1234)
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
	os.Clearenv()
	os.Setenv("WAYBACK_TELEGRAM_CHANNEL", "tg:channel:name")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "tg:channel:name"
	got := opts.TelegramChannel()

	if got != expected {
		t.Fatalf(`Unexpected Telegram channel name, got %v instead of %s`, got, expected)
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
		t.Fatalf(`Unexpected Tor private key, got %v instead of %q`, got, expected)
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
		t.Fatalf(`Unexpected Tor private key, got %v instead of %q`, got, expected)
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
		t.Fatalf(`Unexpected Tor private key, got %v instead of %v`, got, expected)
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
		t.Fatalf(`Unexpected Tor private key, got %v instead of %v`, got, expected)
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
