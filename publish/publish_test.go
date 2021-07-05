// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	telegram "gopkg.in/tucnak/telebot.v2"
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

func unsetAllEnv() {
	lines := os.Environ()
	for _, line := range lines {
		fields := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(fields[0])
		if strings.HasPrefix(key, "WAYBACK_") {
			os.Unsetenv(key)
		}
	}
}

func TestPublishToChannelFromTelegram(t *testing.T) {
	unsetAllEnv()
	setTelegramEnv()
	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()

	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		slug := strings.TrimPrefix(r.URL.Path, prefix)
		switch slug {
		case "getMe":
			fmt.Fprintln(w, getMeJSON)
		case "getChat":
			fmt.Fprintln(w, getChatJSON)
		case "sendMessage":
			text, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(text), config.SlotName(config.SLOT_IA)) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"ok":true, "result":null}`)
		}
	})

	bot, err := telegram.NewBot(telegram.Settings{
		URL:    server.URL,
		Token:  token,
		Client: httpClient,
	})
	if err != nil {
		t.Fatalf(`New Telegram bot API client failed: %v`, err)
	}

	ctx := context.WithValue(context.Background(), FlagTelegram, bot)
	To(ctx, collects, FlagTelegram)
}

func TestPublishTootFromMastodon(t *testing.T) {
	unsetAllEnv()
	setMastodonEnv()

	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer zoo" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		switch r.URL.Path {
		case "/api/v1/statuses":
			status := r.FormValue("status")
			if !strings.Contains(status, config.SlotName(config.SLOT_IA)) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"access_token": "zoo"}`)
		default:
			fmt.Fprintln(w, `{}`)
		}
	})

	os.Setenv("WAYBACK_MASTODON_SERVER", server.URL)

	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()

	mstdn := NewMastodon(nil)

	ctx := context.WithValue(context.Background(), FlagMastodon, mstdn.client)
	To(ctx, collects, FlagMastodon)
}

func TestPublishTweetFromTwitter(t *testing.T) {
	unsetAllEnv()
	setTwitterEnv()
	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()

	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/1.1/statuses/update.json":
			status := r.FormValue("status")
			if !strings.Contains(status, config.SlotName(config.SLOT_IA)) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"id": 1}`)
		default:
			fmt.Fprintln(w, `{}`)
		}
	})

	twi := NewTwitter(twitter.NewClient(httpClient))
	ctx := context.WithValue(context.Background(), FlagTwitter, twi.client)
	To(ctx, collects, FlagTwitter)
}

func TestPublishToIRCChannelFromIRC(t *testing.T) {
	unsetAllEnv()
}

func TestPublishToMatrixRoomFromMatrix(t *testing.T) {
	unsetAllEnv()
	setMatrixEnv()

	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/_matrix/client/r0/login":
			fmt.Fprintln(w, `{"access_token": "zoo"}`)
		case strings.Contains(r.URL.Path, "!bar:example.com/send/m.room.message"):
			body, _ := ioutil.ReadAll(r.Body)
			if !strings.Contains(string(body), config.SlotName(config.SLOT_IA)) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"id": 1}`)
		}
	})

	os.Setenv("WAYBACK_MATRIX_HOMESERVER", server.URL)
	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()

	mat := NewMatrix(nil)
	ctx := context.WithValue(context.Background(), "matrix", mat.client)
	To(ctx, collects, "matrix")
}
