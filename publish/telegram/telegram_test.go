// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/publish/telegram"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"

	telegram "gopkg.in/telebot.v3"
)

var (
	token     = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	prefix    = fmt.Sprintf("/bot%s/", token)
	getMeJSON = `{
  "ok": true,
  "result": {
    "id": 123456,
    "is_bot": true,
    "first_name": "Bot",
    "username": "Fake Bot"
  }
}`
	getChatJSON = `{
	"ok": true,
	"result": {
		"id": -100011121113,
		"title": "Channel Name",
		"username": "channel-id",
		"type": "channel"
	}
}`
)

func setEnv(t *testing.T) {
	t.Setenv("LOG_LEVEL", "fatal")
	t.Setenv("WAYBACK_TELEGRAM_TOKEN", "foo")
	t.Setenv("WAYBACK_TELEGRAM_CHANNEL", "bar")
}

func testServer() (*http.Client, *httptest.Server) {
	httpClient, mux, server := helper.MockServer()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var dat map[string]interface{}
		if err := json.Unmarshal(b, &dat); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		text, _ := dat["text"].(string)

		w.Header().Set("Content-Type", "application/json")
		sub := strings.Split(strings.TrimPrefix(r.URL.Path, prefix), "/")
		slug := sub[len(sub)-1]
		switch slug {
		case "getMe":
			fmt.Fprintln(w, getMeJSON)
		case "getChat":
			fmt.Fprintln(w, getChatJSON)
		case "sendMessage":
			if !strings.Contains(text, config.SlotName(config.SLOT_IA)) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"ok":true, "result":null}`)
		case "sendMediaGroup":
			fmt.Fprintln(w, `{"ok":true, "result":null}`)
		}
	})

	return httpClient, server
}

func TestNew(t *testing.T) {
	t.Setenv("LOG_LEVEL", "fatal")
	t.Setenv("WAYBACK_TELEGRAM_TOKEN", "foo")

	client, server := testServer()
	defer server.Close()

	tests := []struct {
		channel string
		isNil   bool
	}{
		{
			channel: "",
			isNil:   true,
		},
		{
			channel: "bar",
			isNil:   false,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			t.Setenv("WAYBACK_TELEGRAM_CHANNEL", test.channel)
			opts, _ := config.NewParser().ParseEnvironmentVariables()
			actual := New(client, opts) == nil
			if actual != test.isNil {
				t.Errorf(`Unexpected new telegram client, got %v instead of %v`, actual, test.isNil)
			}
		})
	}
}

func TestPublish(t *testing.T) {
	setEnv(t)

	client, server := testServer()
	defer server.Close()

	bot, err := telegram.NewBot(telegram.Settings{
		URL:    server.URL,
		Token:  token,
		Client: client,
	})
	if err != nil {
		t.Fatalf(`New Telegram bot API client failed: %v`, err)
	}

	opts, _ := config.NewParser().ParseEnvironmentVariables()
	tel := &Telegram{bot: bot, opts: opts}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	tests := []struct {
		cols  []wayback.Collect
		isNil bool
	}{
		{
			cols:  []wayback.Collect{},
			isNil: true,
		},
		{
			cols:  publish.Collects,
			isNil: false,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			err := tel.Publish(ctx, reduxer.BundleExample(), test.cols)
			actual := err != nil
			if actual != test.isNil {
				t.Errorf(`Unexpected new telegram client, got %v instead of %v`, actual, test.isNil)
			}
		})
	}
}

func TestShutdown(t *testing.T) {
	setEnv(t)
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	client, server := testServer()
	defer server.Close()

	tel := New(client, opts)
	go tel.bot.Start()
	err := tel.Shutdown()
	if err != nil {
		t.Errorf("Unexpected shutdown: %v", err)
	}
}
