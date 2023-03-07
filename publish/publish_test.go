// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/reduxer"
)

var collects = []wayback.Collect{
	{
		Arc: config.SLOT_IA,
		Dst: "https://web.archive.org/web/20211000000001/https://example.com/",
		Src: "https://example.com/",
		Ext: config.SLOT_IA,
	},
	{
		Arc: config.SLOT_IS,
		Dst: "http://archive.today/abcdE",
		Src: "https://example.com/",
		Ext: config.SLOT_IS,
	},
	{
		Arc: config.SLOT_IP,
		Dst: "https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr",
		Src: "https://example.com/",
		Ext: config.SLOT_IP,
	},
	{
		Arc: config.SLOT_PH,
		Dst: "http://telegra.ph/title-01-01",
		Src: "https://example.com/",
		Ext: config.SLOT_PH,
	},
}

func isBlocking(f func()) bool {
	ch := make(chan struct{})
	go func() {
		f()
		close(ch)
	}()
	select {
	case <-ch:
		return false
	case <-time.After(time.Millisecond * 10):
		return true
	}
}

type mockPublisher struct{}

func (m *mockPublisher) Publish(_ context.Context, _ reduxer.Reduxer, _ []wayback.Collect, args ...string) error {
	return nil
}

func mockRegister(opts *config.Options) *Module {
	publisher := &mockPublisher{}

	return &Module{
		Publisher: publisher,
		Opts:      opts,
	}
}

func TestNew(t *testing.T) {
	defer helper.CheckTest(t)

	ctx := context.Background()
	opts := &config.Options{}

	pub := New(ctx, opts)

	if pub == nil {
		t.Error("Expected non-nil publish service")
	}

	if pub.opts != opts {
		t.Error("Expected publish service options to match input options")
	}

	if pub.pool == nil {
		t.Error("Expected non-nil publish service pool")
	}
}

func TestStart(t *testing.T) {
	defer helper.CheckTest(t)

	pool := &pooling.Pool{}

	pub := &Publish{opts: &config.Options{}, pool: pool}
	go pub.Start()

	// Wait for the Start method to block
	time.Sleep(time.Millisecond * 10)

	// Check if the Start method is blocking
	if !isBlocking(pub.Start) {
		t.Error("Start method is not blocking")
	}
}

func TestStop(t *testing.T) {
	defer helper.CheckTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pub := New(ctx, &config.Options{})
	go pub.Start()

	// Wait a short time to ensure that the service has started
	time.Sleep(time.Millisecond * 100)

	if pub.pool.Closed() {
		t.Errorf("expected publish service to be running, but got %s", pub.pool.Status())
	}

	pub.Stop()

	if !pub.pool.Closed() {
		t.Errorf("expected publish service to be stopped, but got %s", pub.pool.Status())
	}
}

func TestSpread(t *testing.T) {
	defer helper.CheckTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	os.Setenv("WAYBACK_TELEGRAM_TOKEN", "tg:token")
	os.Setenv("WAYBACK_TELEGRAM_CHANNEL", "tg:channel")
	os.Setenv("WAYBACK_GITHUB_REPO", "github-repo")
	os.Setenv("WAYBACK_GITHUB_TOKEN", "github:token")
	os.Setenv("WAYBACK_GITHUB_OWNER", "github-owner")
	os.Setenv("WAYBACK_NOTION_TOKEN", "notion:token")
	os.Setenv("WAYBACK_NOTION_DATABASE_ID", "uuid4")
	os.Setenv("WAYBACK_MASTODON_KEY", "foo")
	os.Setenv("WAYBACK_MASTODON_SECRET", "foo")
	os.Setenv("WAYBACK_MASTODON_TOKEN", "foo")
	os.Setenv("WAYBACK_TWITTER_CONSUMER_KEY", "foo")
	os.Setenv("WAYBACK_TWITTER_CONSUMER_SECRET", "foo")
	os.Setenv("WAYBACK_TWITTER_ACCESS_TOKEN", "foo")
	os.Setenv("WAYBACK_TWITTER_ACCESS_SECRET", "foo")
	os.Setenv("WAYBACK_IRC_NICK", "foo")
	os.Setenv("WAYBACK_IRC_CHANNEL", "bar")
	os.Setenv("WAYBACK_MATRIX_HOMESERVER", "https://matrix-client.matrix.org")
	os.Setenv("WAYBACK_MATRIX_USERID", "@foo:matrix.org")
	os.Setenv("WAYBACK_MATRIX_ROOMID", "!bar:matrix.org")
	os.Setenv("WAYBACK_MATRIX_PASSWORD", "zoo")
	os.Setenv("WAYBACK_DISCORD_BOT_TOKEN", "discord-bot-token")
	os.Setenv("WAYBACK_DISCORD_CHANNEL", "865981235815140000")
	// os.Setenv("WAYBACK_SLACK_APP_TOKEN", "slack-app-token")
	os.Setenv("WAYBACK_SLACK_BOT_TOKEN", "slack-bot-token")
	os.Setenv("WAYBACK_SLACK_CHANNEL", "C123ABCXY89")
	os.Setenv("WAYBACK_NOSTR_RELAY_URL", "wss://nostr.example.com")
	os.Setenv("WAYBACK_NOSTR_PRIVATE_KEY", "nprivabc")
	os.Setenv("WAYBACK_MEILI_ENDPOINT", "https://meilisearch.example.com")

	opts, err := config.NewParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing environment variables failed: %v`, err)
	}

	tests := []struct {
		pubTo Flag
	}{
		{FlagWeb},
		{FlagTelegram},
		{FlagTwitter},
		{FlagMastodon},
		{FlagDiscord},
		{FlagMatrix},
		{FlagSlack},
		{FlagIRC},
		{FlagNotion},
		{FlagMeili},
	}

	pub := New(ctx, opts)
	// go pub.Start()
	// defer pub.Stop()

	for _, test := range tests {
		// Delete exists module to prevent panic.
		if _, exists := modules[test.pubTo]; exists {
			delete(modules, test.pubTo)
		}
		Register(test.pubTo, mockRegister)
	}
	parseModule(opts)

	// Request from Telegram
	pub.Spread(ctx, nil, []wayback.Collect{}, FlagTelegram)

	v := reflect.ValueOf(pub.pool)
	actual := v.Elem().FieldByName("waiting").Int()
	expect := int64(len(tests))
	if actual != expect {
		t.Errorf("unexpected spread to publishers, got %d instead of %d", actual, expect)
	}
}
