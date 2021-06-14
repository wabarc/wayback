// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	telegram "gopkg.in/tucnak/telebot.v2"
)

var message = `<b><a href='https://web.archive.org/'>Internet Archive</a></b>:
• <a href="https://example.com/?q=%E4%B8%AD%E6%96%87">origin</a> - <a href="https://web.archive.org/web/20211000000001/https://example.com/?q=%E4%B8%AD%E6%96%87">https://web.archive.org/web/20211000000001/https://example.com/?q=%E4%B8%AD%E6%96%87</a>

<b><a href='https://archive.today/'>archive.today</a></b>:
• <a href="https://example.com/">origin</a> - <a href="http://archive.today/abcdE">http://archive.today/abcdE</a>

<b><a href='https://ipfs.github.io/public-gateway-checker/'>IPFS</a></b>:
• <a href="https://example.com/">origin</a> - <a href="https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr">https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr</a>

<b><a href='https://telegra.ph/'>Telegraph</a></b>:
• <a href="https://example.com/">origin</a> - <a href="http://telegra.ph/title-01-01">http://telegra.ph/title-01-01</a>
`

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

func TestRenderForTelegram(t *testing.T) {
	setTelegramEnv()

	tel := &Telegram{}
	got := tel.Render(collects)
	if got != message {
		t.Errorf("Unexpected render template for Telegram got \n%s\ninstead of \n%s", got, message)
	}
}

func TestRenderForTelegramFlawed(t *testing.T) {
	setTelegramEnv()

	message := `<b><a href='https://web.archive.org/'>Internet Archive</a></b>:
• <a href="https://example.com/?q=%E4%B8%AD%E6%96%87">origin</a> - Get "https://web.archive.org/save/https://example.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)

<b><a href='https://archive.today/'>archive.today</a></b>:
• <a href="https://example.com/">origin</a> - <a href="http://archive.today/abcdE">http://archive.today/abcdE</a>

<b><a href='https://ipfs.github.io/public-gateway-checker/'>IPFS</a></b>:
• <a href="https://example.com/">origin</a> - Archive failed.

<b><a href='https://telegra.ph/'>Telegraph</a></b>:
• <a href="https://example.com/404">origin</a> - <a href="https://web.archive.org/*/https://webcache.googleusercontent.com/search?q=cache:https://example.com/404">https://web.archive.org/*/https://webcache.googleusercontent.com/search?q=cache:https://example.com/404</a>
`
	tel := &Telegram{}
	got := tel.Render(flawed)
	if got != message {
		t.Errorf("Unexpected render template for Telegram got \n%s\ninstead of \n%s", got, message)
	}
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

	tel := &Telegram{bot: bot}
	got := tel.ToChannel(context.Background(), tel.Render(collects))
	if !got {
		t.Errorf("Unexpected publish Telegram Channel message got %t instead of %t", got, true)
	}
}
