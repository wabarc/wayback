// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/dghubble/go-twitter/twitter"
	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
)

var collects = []*wayback.Collect{
	{
		Arc: config.SlotName(config.SLOT_IA),
		Dst: map[string]string{
			"https://example.com/?q=%E4%B8%AD%E6%96%87": "https://web.archive.org/web/20211000000001/https://example.com/?q=%E4%B8%AD%E6%96%87",
		},
		Ext: config.SlotExtra(config.SLOT_IA),
	},
	{
		Arc: config.SlotName(config.SLOT_IS),
		Dst: map[string]string{
			"https://example.com/": "http://archive.today/abcdE",
		},
		Ext: config.SlotExtra(config.SLOT_IS),
	},
	{
		Arc: config.SlotName(config.SLOT_IP),
		Dst: map[string]string{
			"https://example.com/": "https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr",
		},
		Ext: config.SlotExtra(config.SLOT_IP),
	},
	{
		Arc: config.SlotName(config.SLOT_PH),
		Dst: map[string]string{
			"https://example.com/": "http://telegra.ph/title-01-01",
		},
		Ext: config.SlotExtra(config.SLOT_PH),
	},
}

var flawed = []*wayback.Collect{
	{
		Arc: config.SlotName(config.SLOT_IA),
		Dst: map[string]string{
			"https://example.com/?q=%E4%B8%AD%E6%96%87": `Get "https://web.archive.org/save/https://example.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)`,
		},
		Ext: config.SlotExtra(config.SLOT_IA),
	},
	{
		Arc: config.SlotName(config.SLOT_IS),
		Dst: map[string]string{
			"https://example.com/": "http://archive.today/abcdE",
		},
		Ext: config.SlotExtra(config.SLOT_IS),
	},
	{
		Arc: config.SlotName(config.SLOT_IP),
		Dst: map[string]string{
			"https://example.com/": "Archive failed.",
		},
		Ext: config.SlotExtra(config.SLOT_IP),
	},
	{
		Arc: config.SlotName(config.SLOT_PH),
		Dst: map[string]string{
			"https://example.com/": "Screenshots failed.",
		},
		Ext: config.SlotExtra(config.SLOT_PH),
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
		r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		slug := strings.TrimPrefix(r.URL.Path, prefix)
		switch slug {
		case "getMe":
			fmt.Fprintln(w, getMeJSON)
		case "sendMessage":
			text := r.FormValue("text")
			if strings.Index(text, config.SlotName(config.SLOT_IA)) == -1 {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"ok":true, "result":null}`)
		}
	})

	endpoint := server.URL + "/bot%s/%s"
	bot, err := telegram.NewBotAPIWithClient(token, endpoint, httpClient)
	if err != nil {
		t.Fatalf(`New Telegram bot API client failed: %v`, err)
	}

	ctx := context.WithValue(context.Background(), "telegram", bot)
	To(ctx, collects, "telegram")
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
		r.ParseForm()
		switch r.URL.Path {
		case "/api/v1/statuses":
			status := r.FormValue("status")
			if status != toot {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"access_token": "zoo"}`)
		}
	})

	os.Setenv("WAYBACK_MASTODON_SERVER", server.URL)

	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()

	mstdn := NewMastodon(nil)

	ctx := context.WithValue(context.Background(), "mastodon", mstdn.client)
	To(ctx, collects, "mastodon")
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
			if status != tweet {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"id": 1}`)
		}
	})

	twi := NewTwitter(twitter.NewClient(httpClient))
	ctx := context.WithValue(context.Background(), "twitter", twi.client)
	To(ctx, collects, "twitter")
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
		case strings.Index(r.URL.Path, "!bar:example.com/send/m.room.message") > -1:
			body, _ := ioutil.ReadAll(r.Body)
			if strings.Index(string(body), config.SlotName(config.SLOT_IA)) == -1 {
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
