// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/service/telegram"

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

var (
	times     = 0
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
	getUpdatesJSON = `{
  "ok": true,
  "result": [
    {
      "update_id": 1,
      "message": {
        "message_id": 1001,
        "text": "https://example.com",
        "chat": {
          "id": 1000001,
          "type": "private"
        }
      }
    }
  ]
}`
)

func bot(t *testing.T, done chan<- bool) (*telegram.BotAPI, *httptest.Server) {
	httpClient, mux, server := helper.MockServer()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		slug := strings.TrimPrefix(r.URL.Path, prefix)
		switch slug {
		case "getMe":
			fmt.Fprintln(w, getMeJSON)
		case "getUpdates":
			if times == 0 {
				fmt.Fprintln(w, getUpdatesJSON)
				times++
			} else {
				fmt.Fprintln(w, `{"ok":true, "result":[]}`)
			}
		case "sendMessage":
			text := r.FormValue("text")
			if !strings.Contains(text, config.SlotName("ia")) {
				t.Errorf("Unexpected result: %s", text)
				return
			}
			fmt.Fprintln(w, `{"ok":true, "result":null}`)
			done <- true
		}
	})

	endpoint := server.URL + "/bot%s/%s"
	b, err := telegram.NewBotAPIWithClient(token, endpoint, httpClient)
	if err != nil {
		t.Fatalf(`New Telegram bot API client failed: %v`, err)
	}

	return b, server
}

func TestServe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test in short mode.")
	}

	os.Setenv("WAYBACK_TELEGRAM_TOKEN", token)
	os.Setenv("WAYBACK_TELEGRAM_CHANNEL", "bar")
	os.Setenv("WAYBACK_ENABLE_IA", "true")

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse enviroment variables or flags failed, error: %v", err)
	}

	done := make(chan bool, 1)
	bot, srv := bot(t, done)
	defer srv.Close()

	go func() {
		for {
			select {
			case <-done:
				bot.StopReceivingUpdates()
				return
			case <-time.After(120 * time.Second):
				t.Error("timeout")
				done <- true
			}
		}
	}()

	tg := &Telegram{bot: bot}
	tg.Serve(context.Background())
}

func TestProcess(t *testing.T) {
	t.Skip("Skip")

	os.Setenv("WAYBACK_TELEGRAM_TOKEN", token)
	os.Setenv("WAYBACK_ENABLE_IA", "true")

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse enviroment variables or flags failed, error: %v", err)
	}

	done := make(chan bool, 1)
	bot, srv := bot(t, done)
	defer srv.Close()

	go func() {
		for {
			select {
			case <-done:
				bot.StopReceivingUpdates()
				return
			case <-time.After(120 * time.Second):
				t.Error("timeout")
				done <- true
			}
		}
	}()

	tg := &Telegram{bot: bot}

	cfg := telegram.NewUpdate(0)
	cfg.Timeout = 60
	updates := tg.bot.GetUpdatesChan(cfg)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		t.Log(update.Message.Text)
		if err := tg.process(context.Background(), update); err != nil {
			t.Fatalf("process telegram message failed: %v", err)
		}
	}
}
