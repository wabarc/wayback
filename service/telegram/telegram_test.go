// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/service/telegram"

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
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
	replyJSON = `{
  "ok": true,
  "result": {
    "message_id": 1002,
    "text": "https://example.com",
    "chat": {
      "id": 1000001,
      "type": "private"
    }
  }
}`
)

func handle(mux *http.ServeMux, updatesJSON string) {
	times := 0
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		r.ParseForm()
		text := r.FormValue("text")
		slug := strings.TrimPrefix(r.URL.Path, prefix)
		switch slug {
		case "getMe":
			fmt.Fprintln(w, getMeJSON)
		case "getUpdates":
			if times == 0 {
				fmt.Fprintln(w, updatesJSON)
				times++
			} else {
				fmt.Fprintln(w, `{"ok":true, "result":[]}`)
			}
		case "sendMessage":
			if text == "Archiving..." {
				fmt.Fprintln(w, replyJSON)
				return
			}
			fmt.Fprintln(w, `{"ok":true, "result":null}`)
		case "editMessageText":
			if !strings.Contains(text, config.SlotName("ia")) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"ok":true, "result":null}`)
		case "sendChatAction":
			fmt.Fprintln(w, `{"ok":true, "result":null}`)
		default:
			fmt.Println(slug)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	})
}

func TestServe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test in short mode.")
	}

	os.Setenv("WAYBACK_TELEGRAM_TOKEN", token)
	os.Setenv("WAYBACK_TELEGRAM_CHANNEL", "bar")

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse enviroment variables or flags failed, error: %v", err)
	}

	done := make(chan bool, 1)

	httpClient, mux, server := helper.MockServer()
	defer server.Close()
	handle(mux, `{"ok":true, "result":[]}`)

	endpoint := server.URL + "/bot%s/%s"
	bot, _ := telegram.NewBotAPIWithClient(token, endpoint, httpClient)

	go func() {
		for {
			select {
			case <-done:
				bot.StopReceivingUpdates()
				return
			case <-time.After(3 * time.Second):
				done <- true
			}
		}
	}()

	tg := &Telegram{bot: bot}
	got := tg.Serve(context.Background())
	expected := "done"
	if got.Error() != expected {
		t.Errorf("Unexpected serve telegram got %v instead of %v", got, expected)
	}
}

func TestProcess(t *testing.T) {
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

	httpClient, mux, server := helper.MockServer()
	defer server.Close()
	handle(mux, getUpdatesJSON)

	endpoint := server.URL + "/bot%s/%s"
	bot, _ := telegram.NewBotAPIWithClient(token, endpoint, httpClient)

	go func() {
		for {
			select {
			case <-done:
				bot.StopReceivingUpdates()
				return
			case <-time.After(120 * time.Second):
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

		if err := tg.process(context.Background(), update); err != nil {
			t.Fatalf("process telegram message failed: %v", err)
		} else {
			time.Sleep(3 * time.Second)
			break
		}
	}
	done <- true
}

func TestProcessPlayback(t *testing.T) {
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

	getUpdatesJSON = `{
  "ok": true,
  "result": [
    {
      "update_id": 1,
      "message": {
        "message_id": 1001,
        "text": "/playback https://example.com",
        "entities": [
          {
            "type": "bot_command",
            "offset": 0,
            "length": 9
          }
        ],
        "chat": {
          "id": 1000001,
          "type": "private"
        }
      }
    }
  ]
}`
	httpClient, mux, server := helper.MockServer()
	defer server.Close()
	handle(mux, getUpdatesJSON)

	endpoint := server.URL + "/bot%s/%s"
	bot, _ := telegram.NewBotAPIWithClient(token, endpoint, httpClient)

	go func() {
		for {
			select {
			case <-done:
				bot.StopReceivingUpdates()
				return
			case <-time.After(120 * time.Second):
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

		if err := tg.process(context.Background(), update); err != nil {
			t.Fatalf("process telegram message failed: %v", err)
		} else {
			time.Sleep(time.Second)
			break
		}
	}
	done <- true
}
