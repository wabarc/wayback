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
	"sync"
	"testing"
	"time"

	"github.com/wabarc/logger"
)

func TestSlotName(t *testing.T) {
	tests := []struct {
		slot string
		name string
	}{
		{SLOT_IA, "Internet Archive"},
		{"something", UNKNOWN},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := SlotName(test.slot)

			if got != test.name {
				t.Fatalf(`Unexpected get the slot name description, got %v instead of %s`, got, test.name)
			}
		})
	}
}

func TestSlotExtra(t *testing.T) {
	tests := []struct {
		slot  string
		extra string
	}{
		{SLOT_IA, "https://web.archive.org/"},
		{"something", UNKNOWN},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := SlotExtra(test.slot)

			if got != test.extra {
				t.Errorf(`Unexpected get the slot's extra data, got %v instead of %s`, got, test.extra)
			}
		})
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
	var tests = []struct {
		token      string // managed ipfs token
		userApikey string
		userTarget string
		expected   string
	}{
		{
			token:      "",
			userApikey: "",
			userTarget: "",
			expected:   "",
		},
		{
			token:      "foo",
			userApikey: "",
			userTarget: "",
			expected:   IPFSTarget,
		},
		{
			token:      "",
			userApikey: "bar",
			userTarget: "",
			expected:   "",
		},
		{
			token:      "",
			userApikey: "",
			userTarget: "foo-ipfs-pinning",
			expected:   "foo-ipfs-pinning",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_IPFS_TARGET", test.userTarget)
			os.Setenv("WAYBACK_IPFS_APIKEY", test.userApikey)
			IPFSApikey = test.token

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			expected := test.expected
			got := opts.IPFSTarget()

			if got != expected {
				t.Errorf(`Unexpected IPFS target, got %v instead of %s`, got, expected)
			}
		})
	}
}

func TestIPFSApikey(t *testing.T) {
	var tests = []struct {
		token      string // managed ipfs token
		userApikey string
		userTarget string
		expected   string
	}{
		{
			token:      "",
			userApikey: "",
			userTarget: "",
			expected:   "",
		},
		{
			token:      "foo",
			userApikey: "",
			userTarget: "",
			expected:   "foo",
		},
		{
			token:      "bar",
			userApikey: "zoo",
			userTarget: "",
			expected:   "zoo",
		},
		{
			token:      "",
			userApikey: "",
			userTarget: "foo-ipfs-pinning",
			expected:   "",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_IPFS_TARGET", test.userTarget)
			os.Setenv("WAYBACK_IPFS_APIKEY", test.userApikey)
			IPFSApikey = test.token

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			expected := test.expected
			got := opts.IPFSApikey()

			if got != expected {
				t.Errorf(`Unexpected IPFS apikey, got %v instead of %s`, got, expected)
			}
		})
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
	os.Setenv("WAYBACK_ENABLE_GA", "true")

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
		SLOT_GA: true,
	}
	got := opts.Slots()

	if got == nil || !got[SLOT_IA] || !got[SLOT_IS] || !got[SLOT_IP] || !got[SLOT_PH] || !got[SLOT_GA] {
		t.Fatalf(`Unexpected default slots, got %v instead of %v`, got, expected)
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
		SLOT_IA: true,
		SLOT_IS: true,
		SLOT_IP: true,
		SLOT_PH: true,
		SLOT_GA: true,
	}
	got := opts.Slots()

	if got == nil || !got[SLOT_IA] || !got[SLOT_IS] || !got[SLOT_IP] || !got[SLOT_PH] {
		t.Fatalf(`Unexpected default slots, got %v instead of %v`, got, expected)
	}
}

func TestTelegram(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		call func(*testing.T, *Options, string)
		want string
	}{
		{
			name: "default telegram token",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_TOKEN": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TelegramToken()
				if called != want {
					t.Errorf(`Unexpected get the telegram token, got %v instead of %s`, called, want)
				}
			},
			want: defTelegramToken,
		},
		{
			name: "specified telegram token",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_TOKEN": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TelegramToken()
				if called != want {
					t.Errorf(`Unexpected get the telegram token, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default telegram channel",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_CHANNEL": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TelegramChannel()
				if called != want {
					t.Errorf(`Unexpected get the telegram channel, got %v instead of %s`, called, want)
				}
			},
			want: defTelegramChannel,
		},
		{
			name: "specified telegram channel",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_CHANNEL": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TelegramChannel()
				if called != want {
					t.Errorf(`Unexpected get the telegram channel, got %v instead of %s`, called, want)
				}
			},
			want: "@foo",
		},
		{
			name: "specified telegram channel",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_CHANNEL": "@foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TelegramChannel()
				if called != want {
					t.Errorf(`Unexpected get the telegram channel, got %v instead of %s`, called, want)
				}
			},
			want: "@foo",
		},
		{
			name: "specified telegram channel",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_CHANNEL": "-123456",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TelegramChannel()
				if called != want {
					t.Errorf(`Unexpected get the telegram channel, got %v instead of %s`, called, want)
				}
			},
			want: "-123456",
		},
		{
			name: "default telegram help text",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_HELPTEXT": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TelegramHelptext()
				if called != want {
					t.Errorf(`Unexpected get the telegram help text, got %v instead of %s`, called, want)
				}
			},
			want: defTelegramHelptext,
		},
		{
			name: "specified telegram help text",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_HELPTEXT": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TelegramHelptext()
				if called != want {
					t.Errorf(`Unexpected get the telegram help text, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "publish to telegram enabled",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_TOKEN":   "token",
				"WAYBACK_TELEGRAM_CHANNEL": "@foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToChannel())
				if called != want {
					t.Errorf(`Unexpected enable publish to telegram, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
		{
			name: "publish to telegram disabled",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_TOKEN":   "",
				"WAYBACK_TELEGRAM_CHANNEL": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToChannel())
				if called != want {
					t.Errorf(`Unexpected disable publish to telegram, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to telegram disabled",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_TOKEN":   "",
				"WAYBACK_TELEGRAM_CHANNEL": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToChannel())
				if called != want {
					t.Errorf(`Unexpected disable publish to telegram, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to telegram disabled",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_TOKEN":   "token",
				"WAYBACK_TELEGRAM_CHANNEL": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToChannel())
				if called != want {
					t.Errorf(`Unexpected disable publish to telegram, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "telegram service enabled",
			envs: map[string]string{
				"WAYBACK_TELEGRAM_TOKEN": "token",
			},
			call: func(t *testing.T, opts *Options, want string) {
				opts.EnableServices(ServiceTelegram.String())
				called := strconv.FormatBool(opts.TelegramEnabled())
				if called != want {
					t.Errorf(`Unexpected enable telegram service, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.envs {
				t.Setenv(key, val)
			}
			opts, err := NewParser().ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}
			tt.call(t, opts, tt.want)
		})
	}
}

func TestOnionPrivateKey(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_ONION_PRIVKEY", "onion:private:key")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := "onion:private:key"
	got := opts.OnionPrivKey()

	if got != expected {
		t.Fatalf(`Unexpected Tor private key, got %v instead of %s`, got, expected)
	}
}

func TestOnionLocalPort(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_ONION_LOCAL_PORT", "8080")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := 8080
	got := opts.OnionLocalPort()

	if got != expected {
		t.Fatalf(`Unexpected Tor local port, got %v instead of %q`, got, expected)
	}
}

func TestDefaultOnionLocalPortValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := defOnionLocalPort
	got := opts.OnionLocalPort()

	if got != expected {
		t.Fatalf(`Unexpected Tor local port, got %v instead of %q`, got, expected)
	}
}

func TestTorRemotePorts(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_ONION_REMOTE_PORTS", "80,81,82")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	expected := []int{80, 81, 82}
	got := opts.OnionRemotePorts()

	if got == nil || len(got) != 3 {
		t.Fatalf(`Unexpected Tor remote port, got %v instead of %v`, got, expected)
	}
}

func TestOnionDisabled(t *testing.T) {
	tests := []struct {
		name     string
		disabled bool
		expected bool
	}{
		{"default", defOnionDisabled, false},
		{"disabled", true, true},
		{"enabled", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_ONION_DISABLED", strconv.FormatBool(test.disabled))

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.OnionDisabled()

			if got != test.expected {
				t.Fatalf(`Unexpected disable onion service, got %v instead of %v`, got, test.expected)
			}
		})
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
			os.Setenv("WAYBACK_LISTEN_ADDR", test.address)

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
	got := opts.OnionRemotePorts()

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

func TestPublishToIssues(t *testing.T) {
	os.Clearenv()
	os.Setenv("WAYBACK_GITHUB_REPO", "github-repo")
	os.Setenv("WAYBACK_GITHUB_TOKEN", "github:token")
	os.Setenv("WAYBACK_GITHUB_OWNER", "github-owner")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	ok := opts.PublishToIssues()

	if !ok {
		t.Fatalf(`Unexpected publish to github issue, got %v instead of true`, ok)
	}
}

func TestNotion(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		call func(*testing.T, *Options, string)
		want string
	}{
		{
			name: "default token",
			envs: map[string]string{
				"WAYBACK_NOTION_TOKEN": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.NotionToken()
				if called != want {
					t.Errorf(`Unexpected get the notion token, got %v instead of %s`, called, want)
				}
			},
			want: defNotionToken,
		},
		{
			name: "specified token",
			envs: map[string]string{
				"WAYBACK_NOTION_TOKEN": "foo:bar",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.NotionToken()
				if called != want {
					t.Errorf(`Unexpected get the notion token, got %v instead of %s`, called, want)
				}
			},
			want: "foo:bar",
		},
		{
			name: "default database id",
			envs: map[string]string{
				"WAYBACK_NOTION_DATABASE_ID": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.NotionDatabaseID()
				if called != want {
					t.Errorf(`Unexpected get the database id, got %v instead of %s`, called, want)
				}
			},
			want: defNotionToken,
		},
		{
			name: "specified database id",
			envs: map[string]string{
				"WAYBACK_NOTION_DATABASE_ID": "foo:bar",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.NotionDatabaseID()
				if called != want {
					t.Errorf(`Unexpected get the databases id, got %v instead of %s`, called, want)
				}
			},
			want: "foo:bar",
		},
		{
			name: "publish to notion enabled",
			envs: map[string]string{
				"WAYBACK_NOTION_TOKEN":       "token",
				"WAYBACK_NOTION_DATABASE_ID": "foo-bar-zoo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToNotion())
				if called != want {
					t.Errorf(`Unexpected enable publish to notion, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
		{
			name: "publish to notion disabled",
			envs: map[string]string{
				"WAYBACK_NOTION_TOKEN":       "",
				"WAYBACK_NOTION_DATABASE_ID": "foo-bar-zoo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToNotion())
				if called != want {
					t.Errorf(`Unexpected disable publish to notion, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to notion disabled",
			envs: map[string]string{
				"WAYBACK_NOTION_TOKEN":       "foo",
				"WAYBACK_NOTION_DATABASE_ID": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToNotion())
				if called != want {
					t.Errorf(`Unexpected disable publish to notion, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.envs {
				t.Setenv(key, val)
			}
			opts, err := NewParser().ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}
			tt.call(t, opts, tt.want)
		})
	}
}

func TestMastodon(t *testing.T) {
	server := "https://mastodon.social"
	tests := []struct {
		name string
		envs map[string]string
		call func(*testing.T, *Options, string)
		want string
	}{
		{
			name: "default mastodon server",
			envs: map[string]string{
				"WAYBACK_MASTODON_SERVER": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MastodonServer()
				if called != want {
					t.Errorf(`Unexpected get the mastodon server, got %v instead of %s`, called, want)
				}
			},
			want: defMastodonServer,
		},
		{
			name: "specified mastodon server",
			envs: map[string]string{
				"WAYBACK_MASTODON_SERVER": server,
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MastodonServer()
				if called != want {
					t.Errorf(`Unexpected get the mastodon server, got %v instead of %s`, called, want)
				}
			},
			want: server,
		},
		{
			name: "default mastodon client key",
			envs: map[string]string{
				"WAYBACK_MASTODON_KEY": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MastodonClientKey()
				if called != want {
					t.Errorf(`Unexpected get the mastodon client key, got %v instead of %s`, called, want)
				}
			},
			want: defMastodonClientKey,
		},
		{
			name: "specified mastodon client key",
			envs: map[string]string{
				"WAYBACK_MASTODON_KEY": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MastodonClientKey()
				if called != want {
					t.Errorf(`Unexpected get the mastodon client key, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default mastodon client secret",
			envs: map[string]string{
				"WAYBACK_MASTODON_SECRET": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MastodonClientKey()
				if called != want {
					t.Errorf(`Unexpected get the mastodon access secret, got %v instead of %s`, called, want)
				}
			},
			want: defMastodonClientSecret,
		},
		{
			name: "specified mastodon client secret",
			envs: map[string]string{
				"WAYBACK_MASTODON_SECRET": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MastodonClientSecret()
				if called != want {
					t.Errorf(`Unexpected get the mastodon access secret, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default mastodon access token",
			envs: map[string]string{
				"WAYBACK_MASTODON_TOKEN": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MastodonAccessToken()
				if called != want {
					t.Errorf(`Unexpected get the mastodon access token, got %v instead of %s`, called, want)
				}
			},
			want: defMastodonClientSecret,
		},
		{
			name: "specified mastodon access token",
			envs: map[string]string{
				"WAYBACK_MASTODON_TOKEN": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MastodonAccessToken()
				if called != want {
					t.Errorf(`Unexpected get the mastodon access token, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default mastodon cw",
			envs: map[string]string{
				"WAYBACK_MASTODON_CW": "true",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.MastodonCW())
				if called != want {
					t.Errorf(`Unexpected get the mastodon cw status, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
		{
			name: "specified mastodon cw",
			envs: map[string]string{
				"WAYBACK_MASTODON_CW": "false",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.MastodonCW())
				if called != want {
					t.Errorf(`Unexpected get the mastodon cw status, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "default mastodon cw text",
			envs: map[string]string{
				"WAYBACK_MASTODON_CWTEXT": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				opts.mastodon.cw = true
				called := opts.MastodonCWText()
				if called != want {
					t.Errorf(`Unexpected get the mastodon cw text, got %v instead of %s`, called, want)
				}
			},
			want: defMastodonCWText,
		},
		{
			name: "specified mastodon cw text",
			envs: map[string]string{
				"WAYBACK_MASTODON_CWTEXT": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				opts.mastodon.cw = true
				called := opts.MastodonCWText()
				if called != want {
					t.Errorf(`Unexpected get the mastodon cw text, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "publish to mastodon enabled",
			envs: map[string]string{
				"WAYBACK_MASTODON_KEY":    "foo",
				"WAYBACK_MASTODON_TOKEN":  "foo",
				"WAYBACK_MASTODON_SECRET": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToMastodon())
				if called != want {
					t.Errorf(`Unexpected enable publish to mastodon, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
		{
			name: "publish to mastodon disabled",
			envs: map[string]string{
				"WAYBACK_MASTODON_KEY":    "",
				"WAYBACK_MASTODON_TOKEN":  "foo",
				"WAYBACK_MASTODON_SECRET": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToMastodon())
				if called != want {
					t.Errorf(`Unexpected disable publish to mastodon, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to mastodon disabled",
			envs: map[string]string{
				"WAYBACK_MASTODON_KEY":    "foo",
				"WAYBACK_MASTODON_TOKEN":  "",
				"WAYBACK_MASTODON_SECRET": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToMastodon())
				if called != want {
					t.Errorf(`Unexpected disable publish to mastodon, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to mastodon disabled",
			envs: map[string]string{
				"WAYBACK_MASTODON_KEY":    "foo",
				"WAYBACK_MASTODON_TOKEN":  "foo",
				"WAYBACK_MASTODON_SECRET": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToMastodon())
				if called != want {
					t.Errorf(`Unexpected disable publish to mastodon, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "mastodon service enabled",
			envs: map[string]string{
				"WAYBACK_MASTODON_KEY":    "foo",
				"WAYBACK_MASTODON_TOKEN":  "foo",
				"WAYBACK_MASTODON_SECRET": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				opts.EnableServices(ServiceMastodon.String())
				called := strconv.FormatBool(opts.MastodonEnabled())
				if called != want {
					t.Errorf(`Unexpected enable mastodon service, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.envs {
				t.Setenv(key, val)
			}
			opts, err := NewParser().ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}
			tt.call(t, opts, tt.want)
		})
	}
}

func TestTwitter(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		call func(*testing.T, *Options, string)
		want string
	}{
		{
			name: "default twitter consumer key",
			envs: map[string]string{
				"WAYBACK_TWITTER_CONSUMER_KEY": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TwitterConsumerKey()
				if called != want {
					t.Errorf(`Unexpected get the twitter consumer key, got %v instead of %s`, called, want)
				}
			},
			want: defTwitterConsumerKey,
		},
		{
			name: "specified twitter consumer key",
			envs: map[string]string{
				"WAYBACK_TWITTER_CONSUMER_KEY": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TwitterConsumerKey()
				if called != want {
					t.Errorf(`Unexpected get the twitter consumer key, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default twitter consumer secret",
			envs: map[string]string{
				"WAYBACK_TWITTER_CONSUMER_SECRET": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TwitterConsumerSecret()
				if called != want {
					t.Errorf(`Unexpected get the twitter consumer secret, got %v instead of %s`, called, want)
				}
			},
			want: defTwitterConsumerSecret,
		},
		{
			name: "specified twitter consumer secret",
			envs: map[string]string{
				"WAYBACK_TWITTER_CONSUMER_SECRET": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TwitterConsumerSecret()
				if called != want {
					t.Errorf(`Unexpected get the twitter consumer secret, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default twitter access token",
			envs: map[string]string{
				"WAYBACK_TWITTER_ACCESS_TOKEN": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TwitterAccessToken()
				if called != want {
					t.Errorf(`Unexpected get the twitter access token, got %v instead of %s`, called, want)
				}
			},
			want: defTwitterAccessToken,
		},
		{
			name: "specified twitter access token",
			envs: map[string]string{
				"WAYBACK_TWITTER_ACCESS_TOKEN": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TwitterAccessToken()
				if called != want {
					t.Errorf(`Unexpected get the twitter access token, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default twitter access secret",
			envs: map[string]string{
				"WAYBACK_TWITTER_ACCESS_SECRET": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TwitterAccessSecret()
				if called != want {
					t.Errorf(`Unexpected get the twitter access secret, got %v instead of %s`, called, want)
				}
			},
			want: defTwitterAccessSecret,
		},
		{
			name: "specified twitter access secret",
			envs: map[string]string{
				"WAYBACK_TWITTER_ACCESS_SECRET": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.TwitterAccessSecret()
				if called != want {
					t.Errorf(`Unexpected get the twitter access secret, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "publish to twitter enabled",
			envs: map[string]string{
				"WAYBACK_TWITTER_CONSUMER_KEY":    "foo",
				"WAYBACK_TWITTER_CONSUMER_SECRET": "foo",
				"WAYBACK_TWITTER_ACCESS_TOKEN":    "foo",
				"WAYBACK_TWITTER_ACCESS_SECRET":   "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToTwitter())
				if called != want {
					t.Errorf(`Unexpected disable publish to twitter, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
		{
			name: "publish to twitter disabled",
			envs: map[string]string{
				"WAYBACK_TWITTER_CONSUMER_KEY":    "",
				"WAYBACK_TWITTER_CONSUMER_SECRET": "foo",
				"WAYBACK_TWITTER_ACCESS_TOKEN":    "foo",
				"WAYBACK_TWITTER_ACCESS_SECRET":   "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToTwitter())
				if called != want {
					t.Errorf(`Unexpected disable publish to twitter, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to twitter disabled",
			envs: map[string]string{
				"WAYBACK_TWITTER_CONSUMER_KEY":    "foo",
				"WAYBACK_TWITTER_CONSUMER_SECRET": "",
				"WAYBACK_TWITTER_ACCESS_TOKEN":    "foo",
				"WAYBACK_TWITTER_ACCESS_SECRET":   "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToTwitter())
				if called != want {
					t.Errorf(`Unexpected disable publish to twitter, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to twitter disabled",
			envs: map[string]string{
				"WAYBACK_TWITTER_CONSUMER_KEY":    "foo",
				"WAYBACK_TWITTER_CONSUMER_SECRET": "foo",
				"WAYBACK_TWITTER_ACCESS_TOKEN":    "",
				"WAYBACK_TWITTER_ACCESS_SECRET":   "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToTwitter())
				if called != want {
					t.Errorf(`Unexpected disable publish to twitter, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to twitter disabled",
			envs: map[string]string{
				"WAYBACK_TWITTER_CONSUMER_KEY":    "foo",
				"WAYBACK_TWITTER_CONSUMER_SECRET": "foo",
				"WAYBACK_TWITTER_ACCESS_TOKEN":    "foo",
				"WAYBACK_TWITTER_ACCESS_SECRET":   "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToTwitter())
				if called != want {
					t.Errorf(`Unexpected disable publish to twitter, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "twitter service enabled",
			envs: map[string]string{
				"WAYBACK_TWITTER_CONSUMER_KEY":    "foo",
				"WAYBACK_TWITTER_CONSUMER_SECRET": "foo",
				"WAYBACK_TWITTER_ACCESS_TOKEN":    "foo",
				"WAYBACK_TWITTER_ACCESS_SECRET":   "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				opts.EnableServices(ServiceTwitter.String())
				called := strconv.FormatBool(opts.TwitterEnabled())
				if called != want {
					t.Errorf(`Unexpected enable twitter service, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.envs {
				t.Setenv(key, val)
			}
			opts, err := NewParser().ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}
			tt.call(t, opts, tt.want)
		})
	}
}

func TestIRC(t *testing.T) {
	server := "example.com:7700"
	tests := []struct {
		name string
		envs map[string]string
		call func(*testing.T, *Options, string)
		want string
	}{
		{
			name: "default irc nick",
			envs: map[string]string{
				"WAYBACK_IRC_NICK": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.IRCNick()
				if called != want {
					t.Errorf(`Unexpected get the irc nick, got %v instead of %s`, called, want)
				}
			},
			want: defIRCNick,
		},
		{
			name: "specified irc nick",
			envs: map[string]string{
				"WAYBACK_IRC_NICK": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.IRCNick()
				if called != want {
					t.Errorf(`Unexpected get the irc nick, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default irc name",
			envs: map[string]string{
				"WAYBACK_IRC_NAME": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.IRCName()
				if called != want {
					t.Errorf(`Unexpected get the irc name, got %v instead of %s`, called, want)
				}
			},
			want: defIRCName,
		},
		{
			name: "specified irc name",
			envs: map[string]string{
				"WAYBACK_IRC_NAME": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.IRCName()
				if called != want {
					t.Errorf(`Unexpected get the irc name, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default irc password",
			envs: map[string]string{
				"WAYBACK_IRC_PASSWORD": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.IRCPassword()
				if called != want {
					t.Errorf(`Unexpected get the irc password, got %v instead of %s`, called, want)
				}
			},
			want: defIRCPassword,
		},
		{
			name: "specified irc password",
			envs: map[string]string{
				"WAYBACK_IRC_PASSWORD": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.IRCPassword()
				if called != want {
					t.Errorf(`Unexpected get the irc password, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default irc channel",
			envs: map[string]string{
				"WAYBACK_IRC_CHANNEL": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.IRCChannel()
				if called != want {
					t.Errorf(`Unexpected get the irc channel, got %v instead of %s`, called, want)
				}
			},
			want: defIRCChannel,
		},
		{
			name: "specified irc channel",
			envs: map[string]string{
				"WAYBACK_IRC_CHANNEL": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.IRCChannel()
				if called != want {
					t.Errorf(`Unexpected get the irc channel, got %v instead of %s`, called, want)
				}
			},
			want: "#foo",
		},
		{
			name: "specified irc channel with prefix",
			envs: map[string]string{
				"WAYBACK_IRC_CHANNEL": "#foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.IRCChannel()
				if called != want {
					t.Errorf(`Unexpected get the irc channel, got %v instead of %s`, called, want)
				}
			},
			want: "#foo",
		},
		{
			name: "default irc server",
			envs: map[string]string{
				"WAYBACK_IRC_SERVER": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.IRCServer()
				if called != want {
					t.Errorf(`Unexpected get the irc server, got %v instead of %s`, called, want)
				}
			},
			want: defIRCServer,
		},
		{
			name: "specified irc server",
			envs: map[string]string{
				"WAYBACK_IRC_SERVER": server,
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.IRCServer()
				if called != want {
					t.Errorf(`Unexpected get the irc server, got %v instead of %s`, called, want)
				}
			},
			want: server,
		},
		{
			name: "publish to irc channel enabled",
			envs: map[string]string{
				"WAYBACK_IRC_NICK":    "foo",
				"WAYBACK_IRC_CHANNEL": "bar",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToIRCChannel())
				if called != want {
					t.Errorf(`Unexpected enable publish to irc channel, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
		{
			name: "publish to irc channel disabled",
			envs: map[string]string{
				"WAYBACK_IRC_NICK":    "",
				"WAYBACK_IRC_CHANNEL": "bar",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToIRCChannel())
				if called != want {
					t.Errorf(`Unexpected disable publish to irc channel, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to irc channel disabled",
			envs: map[string]string{
				"WAYBACK_IRC_NICK":    "foo",
				"WAYBACK_IRC_CHANNEL": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToIRCChannel())
				if called != want {
					t.Errorf(`Unexpected disable publish to irc channel, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "irc service enabled",
			envs: map[string]string{
				"WAYBACK_IRC_NICK": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				opts.EnableServices(ServiceIRC.String())
				called := strconv.FormatBool(opts.IRCEnabled())
				if called != want {
					t.Errorf(`Unexpected enable irc service, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.envs {
				t.Setenv(key, val)
			}
			opts, err := NewParser().ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}
			tt.call(t, opts, tt.want)
		})
	}
}

func TestMatrix(t *testing.T) {
	server := "https://matrix-client.matrix.org"
	userid := "@foo:matrix.org"
	roomid := "!foo:matrix.org"
	tests := []struct {
		name string
		envs map[string]string
		call func(*testing.T, *Options, string)
		want string
	}{
		{
			name: "default matrix home server",
			envs: map[string]string{
				"WAYBACK_MATRIX_HOMESERVER": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MatrixHomeserver()
				if called != want {
					t.Errorf(`Unexpected get the matrix homeserver, got %v instead of %s`, called, want)
				}
			},
			want: defMatrixHomeserver,
		},
		{
			name: "specified matrix home server",
			envs: map[string]string{
				"WAYBACK_MATRIX_HOMESERVER": server,
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MatrixHomeserver()
				if called != want {
					t.Errorf(`Unexpected get the matrix homeserver, got %v instead of %s`, called, want)
				}
			},
			want: server,
		},
		{
			name: "default matrix user id",
			envs: map[string]string{
				"WAYBACK_MATRIX_USERID": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MatrixUserID()
				if called != want {
					t.Errorf(`Unexpected get the matrix user id, got %v instead of %s`, called, want)
				}
			},
			want: defMatrixUserID,
		},
		{
			name: "specified matrix user id",
			envs: map[string]string{
				"WAYBACK_MATRIX_USERID": userid,
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MatrixUserID()
				if called != want {
					t.Errorf(`Unexpected get the matrix user id, got %v instead of %s`, called, want)
				}
			},
			want: userid,
		},
		{
			name: "default matrix room id",
			envs: map[string]string{
				"WAYBACK_MATRIX_ROOMID": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MatrixRoomID()
				if called != want {
					t.Errorf(`Unexpected get the matrix room id, got %v instead of %s`, called, want)
				}
			},
			want: defMatrixRoomID,
		},
		{
			name: "specified matrix room id",
			envs: map[string]string{
				"WAYBACK_MATRIX_ROOMID": roomid,
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MatrixRoomID()
				if called != want {
					t.Errorf(`Unexpected get the matrix room id, got %v instead of %s`, called, want)
				}
			},
			want: roomid,
		},
		{
			name: "default matrix password",
			envs: map[string]string{
				"WAYBACK_MATRIX_PASSWORD": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MatrixPassword()
				if called != want {
					t.Errorf(`Unexpected get the matrix password, got %v instead of %s`, called, want)
				}
			},
			want: defMatrixPassword,
		},
		{
			name: "specified matrix password",
			envs: map[string]string{
				"WAYBACK_MATRIX_PASSWORD": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.MatrixPassword()
				if called != want {
					t.Errorf(`Unexpected get the matrix password, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "publish to matrix enabled",
			envs: map[string]string{
				"WAYBACK_MATRIX_USERID":   "@foo:matrix.org",
				"WAYBACK_MATRIX_ROOMID":   "!bar:matrix.org",
				"WAYBACK_MATRIX_PASSWORD": "zoo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToMatrixRoom())
				if called != want {
					t.Errorf(`Unexpected enable publish to matrix channel, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
		{
			name: "publish to matrix disabled",
			envs: map[string]string{
				"WAYBACK_MATRIX_USERID":   "",
				"WAYBACK_MATRIX_ROOMID":   "!bar:matrix.org",
				"WAYBACK_MATRIX_PASSWORD": "zoo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToMatrixRoom())
				if called != want {
					t.Errorf(`Unexpected disable publish to matrix channel, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to matrix disabled",
			envs: map[string]string{
				"WAYBACK_MATRIX_USERID":   "@foo:matrix.org",
				"WAYBACK_MATRIX_ROOMID":   "",
				"WAYBACK_MATRIX_PASSWORD": "zoo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToMatrixRoom())
				if called != want {
					t.Errorf(`Unexpected disable publish to matrix channel, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to matrix disabled",
			envs: map[string]string{
				"WAYBACK_MATRIX_USERID":   "@foo:matrix.org",
				"WAYBACK_MATRIX_ROOMID":   "!bar:matrix.org",
				"WAYBACK_MATRIX_PASSWORD": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToMatrixRoom())
				if called != want {
					t.Errorf(`Unexpected disable publish to matrix channel, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "matrix service enabled",
			envs: map[string]string{
				"WAYBACK_MATRIX_USERID":   "@foo:matrix.org",
				"WAYBACK_MATRIX_PASSWORD": "zoo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				opts.EnableServices(ServiceMatrix.String())
				called := strconv.FormatBool(opts.MatrixEnabled())
				if called != want {
					t.Errorf(`Unexpected enable matrix service, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.envs {
				t.Setenv(key, val)
			}
			opts, err := NewParser().ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}
			tt.call(t, opts, tt.want)
		})
	}
}

func TestDiscord(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		call func(*testing.T, *Options, string)
		want string
	}{
		{
			name: "default discord bot token",
			envs: map[string]string{
				"WAYBACK_DISCORD_BOT_TOKEN": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.DiscordBotToken()
				if called != want {
					t.Errorf(`Unexpected get the discord bot token, got %v instead of %s`, called, want)
				}
			},
			want: defDiscordBotToken,
		},
		{
			name: "specified discord bot token",
			envs: map[string]string{
				"WAYBACK_DISCORD_BOT_TOKEN": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.DiscordBotToken()
				if called != want {
					t.Errorf(`Unexpected get the discord bot token, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default discord channel",
			envs: map[string]string{
				"WAYBACK_DISCORD_CHANNEL": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.DiscordChannel()
				if called != want {
					t.Errorf(`Unexpected get the discord channel, got %v instead of %s`, called, want)
				}
			},
			want: defDiscordChannel,
		},
		{
			name: "specified discord channel",
			envs: map[string]string{
				"WAYBACK_DISCORD_CHANNEL": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.DiscordChannel()
				if called != want {
					t.Errorf(`Unexpected get the discord channel, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default discord help text",
			envs: map[string]string{
				"WAYBACK_DISCORD_HELPTEXT": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.DiscordHelptext()
				if called != want {
					t.Errorf(`Unexpected get the discord help text, got %v instead of %s`, called, want)
				}
			},
			want: defDiscordHelptext,
		},
		{
			name: "specified discord help text",
			envs: map[string]string{
				"WAYBACK_DISCORD_HELPTEXT": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.DiscordHelptext()
				if called != want {
					t.Errorf(`Unexpected get the discord help text, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "publish discord enabled",
			envs: map[string]string{
				"WAYBACK_DISCORD_BOT_TOKEN": "foo",
				"WAYBACK_DISCORD_CHANNEL":   "bar",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToDiscordChannel())
				if called != want {
					t.Errorf(`Unexpected enable publish to discord channel, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
		{
			name: "publish discord disabled",
			envs: map[string]string{
				"WAYBACK_DISCORD_BOT_TOKEN": "",
				"WAYBACK_DISCORD_CHANNEL":   "bar",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToDiscordChannel())
				if called != want {
					t.Errorf(`Unexpected disable publish to discord channel, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish discord disabled",
			envs: map[string]string{
				"WAYBACK_DISCORD_BOT_TOKEN": "foo",
				"WAYBACK_DISCORD_CHANNEL":   "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToDiscordChannel())
				if called != want {
					t.Errorf(`Unexpected disable publish to discord channel, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "discord service enabled",
			envs: map[string]string{
				"WAYBACK_DISCORD_BOT_TOKEN": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				opts.EnableServices(ServiceDiscord.String())
				called := strconv.FormatBool(opts.DiscordEnabled())
				if called != want {
					t.Errorf(`Unexpected enable discord service, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.envs {
				t.Setenv(key, val)
			}
			opts, err := NewParser().ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}
			tt.call(t, opts, tt.want)
		})
	}
}

func TestSlack(t *testing.T) {
	botToken := "xoxb-2306408000000-2300127000000-GgLHgzqK3fXH5KA50AAbcdef"
	appToken := "xapp-1-A0000000FC7-2300600000035-a000000bc7d104f053f66000000e589dafabcde70c5152abaacbcaea00000000"
	tests := []struct {
		name string
		envs map[string]string
		call func(*testing.T, *Options, string)
		want string
	}{
		{
			name: "default slack bot token",
			envs: map[string]string{
				"WAYBACK_SLACK_BOT_TOKEN": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.SlackBotToken()
				if called != want {
					t.Errorf(`Unexpected get the slack bot token, got %v instead of %s`, called, want)
				}
			},
			want: defSlackBotToken,
		},
		{
			name: "specified slack bot token",
			envs: map[string]string{
				"WAYBACK_SLACK_BOT_TOKEN": botToken,
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.SlackBotToken()
				if called != want {
					t.Errorf(`Unexpected get the slack bot token, got %v instead of %s`, called, want)
				}
			},
			want: botToken,
		},
		{
			name: "default slack app token",
			envs: map[string]string{
				"WAYBACK_SLACK_APP_TOKEN": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.SlackAppToken()
				if called != want {
					t.Errorf(`Unexpected get the slack app token, got %v instead of %s`, called, want)
				}
			},
			want: defSlackAppToken,
		},
		{
			name: "specified slack app token",
			envs: map[string]string{
				"WAYBACK_SLACK_APP_TOKEN": appToken,
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.SlackAppToken()
				if called != want {
					t.Errorf(`Unexpected get the slack app token, got %v instead of %s`, called, want)
				}
			},
			want: appToken,
		},
		{
			name: "default slack channel",
			envs: map[string]string{
				"WAYBACK_SLACK_CHANNEL": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.SlackChannel()
				if called != want {
					t.Errorf(`Unexpected get the slack channel, got %v instead of %s`, called, want)
				}
			},
			want: defSlackChannel,
		},
		{
			name: "specified slack channel",
			envs: map[string]string{
				"WAYBACK_SLACK_CHANNEL": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.SlackChannel()
				if called != want {
					t.Errorf(`Unexpected get the slack channel, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default slack help text",
			envs: map[string]string{
				"WAYBACK_SLACK_HELPTEXT": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.SlackHelptext()
				if called != want {
					t.Errorf(`Unexpected get the slack help text, got %v instead of %s`, called, want)
				}
			},
			want: defSlackHelptext,
		},
		{
			name: "specified slack help text",
			envs: map[string]string{
				"WAYBACK_SLACK_HELPTEXT": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.SlackHelptext()
				if called != want {
					t.Errorf(`Unexpected get the slack help text, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "publish to slack enabled",
			envs: map[string]string{
				"WAYBACK_SLACK_BOT_TOKEN": "foo",
				"WAYBACK_SLACK_CHANNEL":   "bar",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToSlackChannel())
				if called != want {
					t.Errorf(`Unexpected enable publish to slack channel, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
		{
			name: "publish to slack disabled",
			envs: map[string]string{
				"WAYBACK_SLACK_BOT_TOKEN": "",
				"WAYBACK_SLACK_CHANNEL":   "bar",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToSlackChannel())
				if called != want {
					t.Errorf(`Unexpected disable publish to slack channel, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "publish to slack disabled",
			envs: map[string]string{
				"WAYBACK_SLACK_BOT_TOKEN": "foo",
				"WAYBACK_SLACK_CHANNEL":   "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.PublishToSlackChannel())
				if called != want {
					t.Errorf(`Unexpected disable publish to slack channel, got %v instead of %s`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "slack service enabled",
			envs: map[string]string{
				"WAYBACK_SLACK_BOT_TOKEN": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				opts.EnableServices(ServiceSlack.String())
				called := strconv.FormatBool(opts.SlackEnabled())
				if called != want {
					t.Errorf(`Unexpected enable slack service, got %v instead of %s`, called, want)
				}
			},
			want: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.envs {
				t.Setenv(key, val)
			}
			opts, err := NewParser().ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}
			tt.call(t, opts, tt.want)
		})
	}
}

func TestXMPP(t *testing.T) {
	username := "foo@example.com"
	tests := []struct {
		name string
		envs map[string]string
		call func(*testing.T, *Options, string)
		want string
	}{
		{
			name: "default xmpp username",
			envs: map[string]string{
				"WAYBACK_XMPP_USERNAME": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.XMPPUsername()
				if called != want {
					t.Errorf(`Unexpected XMPP username got %v instead of %v`, called, want)
				}
			},
			want: defXMPPUsername,
		},
		{
			name: "specified xmpp username",
			envs: map[string]string{
				"WAYBACK_XMPP_USERNAME": username,
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.XMPPUsername()
				if called != want {
					t.Errorf(`Unexpected XMPP username got %v instead of %v`, called, want)
				}
			},
			want: username,
		},
		{
			name: "default xmpp password",
			envs: map[string]string{
				"WAYBACK_XMPP_PASSWORD": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.XMPPPassword()
				if called != want {
					t.Errorf(`Unexpected XMPP password got %v instead of %v`, called, want)
				}
			},
			want: defXMPPPassword,
		},
		{
			name: "specified xmpp password",
			envs: map[string]string{
				"WAYBACK_XMPP_PASSWORD": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.XMPPPassword()
				if called != want {
					t.Errorf(`Unexpected XMPP password got %v instead of %v`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "default xmpp notls",
			envs: map[string]string{
				"WAYBACK_XMPP_NOTLS": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.XMPPNoTLS())
				if called != want {
					t.Errorf(`Unexpected XMPP password got %v instead of %v`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "specified xmpp notls",
			envs: map[string]string{
				"WAYBACK_XMPP_NOTLS": "true",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.XMPPNoTLS())
				if called != want {
					t.Errorf(`Unexpected XMPP password got %v instead of %v`, called, want)
				}
			},
			want: "true",
		},
		{
			name: "default xmpp help text",
			envs: map[string]string{
				"WAYBACK_XMPP_HELPTEXT": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.XMPPHelptext()
				if called != want {
					t.Errorf(`Unexpected get the xmpp help text, got %v instead of %s`, called, want)
				}
			},
			want: defXMPPHelptext,
		},
		{
			name: "specified xmpp help text",
			envs: map[string]string{
				"WAYBACK_XMPP_HELPTEXT": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := opts.XMPPHelptext()
				if called != want {
					t.Errorf(`Unexpected get the xmpp help text, got %v instead of %s`, called, want)
				}
			},
			want: "foo",
		},
		{
			name: "xmpp service enabled",
			envs: map[string]string{
				"WAYBACK_XMPP_USERNAME": "foo",
				"WAYBACK_XMPP_PASSWORD": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				opts.EnableServices(ServiceXMPP.String())
				called := strconv.FormatBool(opts.XMPPEnabled())
				if called != want {
					t.Errorf(`Unexpected enable xmpp service got %v instead of %v`, called, want)
				}
			},
			want: "true",
		},
		{
			name: "xmpp service disabled",
			envs: map[string]string{
				"WAYBACK_XMPP_USERNAME": "",
				"WAYBACK_XMPP_PASSWORD": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.XMPPEnabled())
				if called != want {
					t.Errorf(`Unexpected disable xmpp service got %v instead of %v`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "xmpp service disabled",
			envs: map[string]string{
				"WAYBACK_XMPP_USERNAME": "",
				"WAYBACK_XMPP_PASSWORD": "foo",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.XMPPEnabled())
				if called != want {
					t.Errorf(`Unexpected disable xmpp service got %v instead of %v`, called, want)
				}
			},
			want: "false",
		},
		{
			name: "xmpp service disabled",
			envs: map[string]string{
				"WAYBACK_XMPP_USERNAME": "foo",
				"WAYBACK_XMPP_PASSWORD": "",
			},
			call: func(t *testing.T, opts *Options, want string) {
				called := strconv.FormatBool(opts.XMPPEnabled())
				if called != want {
					t.Errorf(`Unexpected disable xmpp service got %v instead of %v`, called, want)
				}
			},
			want: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.envs {
				t.Setenv(key, val)
			}
			opts, err := NewParser().ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}
			tt.call(t, opts, tt.want)
		})
	}
}

func TestNostrRelayURL(t *testing.T) {
	var tests = []struct {
		url string
		exp []string
	}{
		{
			url: "",
			exp: []string{defNostrRelayURL},
		},
		{
			url: "wss://example.com",
			exp: []string{"wss://example.com"},
		},
		{
			url: "example.com",
			exp: []string{"example.com"},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_NOSTR_RELAY_URL", test.url)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.NostrRelayURL()
			if len(got) < 1 {
				t.Fatalf(`Unexpected set nostr relay url, got %v instead of %v`, got, test.exp)
			}
			if got[0] != test.exp[0] {
				t.Fatalf(`Unexpected set nostr relay url, got %v instead of %v`, got, test.exp)
			}
		})
	}
}

func TestNostrPrivateKey(t *testing.T) {
	var tests = []struct {
		sk  string
		exp string
	}{
		{
			sk:  "",
			exp: defNostrPrivateKey,
		},
		{
			sk:  "nsecba",
			exp: "nsecba",
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_NOSTR_PRIVATE_KEY", test.sk)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.NostrPrivateKey()
			if got != test.exp {
				t.Fatalf(`Unexpected set nostr private key, got %v instead of %v`, got, test.exp)
			}
		})
	}
}

func TestPublishToNostr(t *testing.T) {
	var tests = []struct {
		sk  string
		url string
		exp bool
	}{
		{
			sk:  "",
			url: "",
			exp: false,
		},
		{
			sk:  "nsecba",
			url: "",
			exp: true,
		},
		{
			sk:  "",
			url: "example.com",
			exp: false,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_NOSTR_RELAY_URL", test.url)
			os.Setenv("WAYBACK_NOSTR_PRIVATE_KEY", test.sk)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.PublishToNostr()
			if got != test.exp {
				t.Fatalf(`Unexpected set nostr private key, got %v instead of %v`, got, test.exp)
			}
		})
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
	var tests = []struct {
		dir string
		exp string
	}{
		{
			dir: "",
			exp: defStorageDir,
		},
		{
			dir: "/path/to/storage",
			exp: "/path/to/storage",
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

			got := opts.StorageDir()
			if got != test.exp {
				t.Errorf(`Unexpected storage binary directory got %s instead of %s`, got, test.dir)
			}
		})
	}
}

func TestEnabledReduxer(t *testing.T) {
	var tests = []struct {
		dir string
		exp bool
	}{
		{
			dir: "",
			exp: true,
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

func TestMaxAttachSize(t *testing.T) {
	parser := NewParser()
	opts, _ := parser.ParseEnvironmentVariables()
	got := opts.MaxAttachSize("telegram")
	if got != maxAttachSizeTelegram {
		t.Fatalf(`Unexpected set wayback timeout got %d instead of %d`, got, maxAttachSizeTelegram)
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
			expected: defMeiliEndpoint,
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

			got := opts.MeiliEndpoint()
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
			expected: defMeiliIndexing,
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

			got := opts.MeiliIndexing()
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
			expected: defMeiliApikey,
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

			got := opts.MeiliApikey()
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

func TestOmnivoreApikey(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		apikey   string
		expected string
	}{
		{
			apikey:   "",
			expected: defOmnivoreApikey,
		},
		{
			apikey:   "foo.bar",
			expected: "foo.bar",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_OMNIVORE_APIKEY", test.apikey)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.OmnivoreApikey()
			if got != test.expected {
				t.Fatalf(`Unexpected set Omnivore api key got %s instead of %s`, got, test.expected)
			}
		})
	}
}

func TestEnabledOmnivore(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		apikey   string
		expected bool
	}{
		{
			apikey:   "",
			expected: false,
		},
		{
			apikey:   "foo-bar",
			expected: true,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_OMNIVORE_APIKEY", test.apikey)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.EnabledOmnivore()
			if got != test.expected {
				t.Fatalf(`Unexpected enabled meilisearch got %t instead of %t`, got, test.expected)
			}
		})
	}
}

func TestEnableServices(t *testing.T) {
	tests := []struct {
		name     string
		services []string
	}{
		{
			name:     "enable Discord service",
			services: []string{"discord"},
		},
		{
			name:     "enable HTTPd service",
			services: []string{"httpd", "web"},
		},
		{
			name:     "enable Mastodon service",
			services: []string{"mastodon", "mstdn"},
		},
		{
			name:     "enable Matrix service",
			services: []string{"matrix"},
		},
		{
			name:     "enable IRC service",
			services: []string{"irc"},
		},
		{
			name:     "enable Slack service",
			services: []string{"slack"},
		},
		{
			name:     "enable Telegram service",
			services: []string{"telegram"},
		},
		{
			name:     "enable Twitter service",
			services: []string{"twitter"},
		},
		{
			name:     "enable XMPP service",
			services: []string{"xmpp"},
		},
		{
			name:     "enable multiple services",
			services: []string{"discord", "httpd", "matrix"},
		},
		{
			name:     "enable unknown services",
			services: []string{"unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{services: sync.Map{}}
			opts.EnableServices(tt.services...)
		})
	}
}

func TestProxy(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		address  string
		expected string
	}{
		{
			address:  "",
			expected: defProxy,
		},
		{
			address:  "http://127.0.0.1",
			expected: `http://127.0.0.1`,
		},
		{
			address:  "http://127.0.0.1:1080",
			expected: `http://127.0.0.1:1080`,
		},
		{
			address:  "https://127.0.0.1:1080",
			expected: `https://127.0.0.1:1080`,
		},
		{
			address:  "socks5://127.0.0.1:1080",
			expected: `socks5://127.0.0.1:1080`,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WAYBACK_PROXY", test.address)

			parser := NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf(`Parsing environment variables failed: %v`, err)
			}

			got := opts.Proxy()
			if got != test.expected {
				t.Fatalf(`Unexpected get proxy, got %s instead of %s`, got, test.expected)
			}
		})
	}
}
