// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

var message = `<b><a href='https://web.archive.org/'>Internet Archive</a></b>:
• <a href="https://example.com/?q=%E4%B8%AD%E6%96%87">origin</a> - https://web.archive.org/web/20211000000001/https://example.com/?q=%E4%B8%AD%E6%96%87

<b><a href='https://archive.today/'>archive.today</a></b>:
• <a href="https://example.com/">origin</a> - http://archive.today/abcdE

<b><a href='https://ipfs.github.io/public-gateway-checker/'>IPFS</a></b>:
• <a href="https://example.com/">origin</a> - https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr

<b><a href='https://telegra.ph/'>Telegraph</a></b>:
• <a href="https://example.com/">origin</a> - http://telegra.ph/title-01-01
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

func TestToChannel(t *testing.T) {
	setTelegramEnv()

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

	tel := &Telegram{bot: bot}
	got := tel.ToChannel(context.Background(), tel.Render(collects))
	if !got {
		t.Errorf("Unexpected publish Telegram Channel message got %t instead of %t", got, true)
	}
}
