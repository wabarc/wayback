// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"

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

func setTelegramEnv() {
	os.Setenv("WAYBACK_TELEGRAM_TOKEN", "foo")
	os.Setenv("WAYBACK_TELEGRAM_CHANNEL", "bar")

	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()
}

func TestToChannel(t *testing.T) {
	setTelegramEnv()

	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var dat map[string]interface{}
		if err := json.Unmarshal(b, &dat); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		text, _ := dat["text"].(string)

		w.Header().Set("Content-Type", "application/json")
		slug := strings.TrimPrefix(r.URL.Path, prefix)
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

	bot, err := telegram.NewBot(telegram.Settings{
		URL:    server.URL,
		Token:  token,
		Client: httpClient,
	})
	if err != nil {
		t.Fatalf(`New Telegram bot API client failed: %v`, err)
	}

	tel := &telegramBot{bot: bot}
	txt := render.ForPublish(&render.Telegram{Cols: collects}).String()
	got := tel.toChannel(reduxer.Artifact{}, "", txt)
	if !got {
		t.Errorf("Unexpected publish Telegram Channel message got %t instead of %t", got, true)
	}
}
