// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package telegram // import "github.com/wabarc/wayback/service/telegram"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/storage"
	telegram "gopkg.in/tucnak/telebot.v2"
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
	getMyCommandsJSON = `{
  "ok": true,
  "result": [
    {
      "command": "help",
      "description": "Show help information"
    },
    {
      "command": "metrics",
      "description": "Show service metrics"
    },
    {
      "command": "playback",
      "description": "Playback archived url"
    }
  ]
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
    "text": "Queue... or Archiving...",
    "from": {
      "id": 120000000,
      "is_bot": true,
      "first_name": "Testing Bot",
      "username": "username"
    },
    "chat": {
      "id": 1000001,
      "type": "private"
    },
    "reply_to_message": {
      "message_id": 1001,
      "text": "https://example.com",
      "chat": {
        "id": 1000001,
        "type": "private"
      }
    }
  }
}`
	sendMessageJSON = `{
  "ok": true,
  "result": {
    "message_id": 1002,
    "text": "message content",
    "from": {
      "id": 120000000,
      "is_bot": true,
      "first_name": "Testing Bot",
      "username": "username"
    },
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

		b, _ := io.ReadAll(r.Body)
		var dat map[string]interface{}
		if err := json.Unmarshal(b, &dat); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		text, _ := dat["text"].(string)

		slug := strings.TrimPrefix(r.URL.Path, prefix)
		switch slug {
		case "getMe":
			fmt.Fprintln(w, getMeJSON)
		case "getChat":
			fmt.Fprintln(w, getChatJSON)
		case "getMyCommands":
			fmt.Fprintln(w, getMyCommandsJSON)
		case "setMyCommands":
			fmt.Fprintln(w, `{"ok":true, "result":true}`)
		case "getUpdates":
			if times == 0 {
				fmt.Fprintln(w, updatesJSON)
				times++
			} else {
				fmt.Fprintln(w, `{"ok":true, "result":[]}`)
			}
		case "sendMessage":
			if text == "Queue..." || strings.Contains(text, config.SlotName("ia")) {
				fmt.Fprintln(w, replyJSON)
			} else {
				fmt.Fprintln(w, sendMessageJSON)
			}
		case "editMessageText":
			if strings.Contains(text, config.SlotName("ia")) || strings.Contains(text, "Archiving...") {
				fmt.Fprintln(w, replyJSON)
			} else {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			}
		case "sendChatAction", "sendMediaGroup":
			fmt.Fprintln(w, `{"ok":true, "result":null}`)
		default:
			fmt.Println(slug)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	})
}

func newTelegram(client *http.Client, endpoint string) (tg *Telegram, cancel context.CancelFunc, err error) {
	bot, err := telegram.NewBot(telegram.Settings{
		URL:    endpoint,
		Token:  token,
		Client: client,
		Poller: &telegram.LongPoller{Timeout: time.Second},
	})
	if err != nil {
		return tg, nil, err
	}

	store, e := storage.Open("")
	if e != nil {
		return tg, nil, e
	}
	ctx, cancel := context.WithCancel(context.Background())
	pool := pooling.New(config.Opts.PoolingSize())
	go func() {
		select {
		case <-ctx.Done():
			pool.Close()
		}
	}()

	tg = &Telegram{ctx: ctx, bot: bot, pool: pool, store: store}

	return tg, cancel, nil
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
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	httpClient, mux, server := helper.MockServer()
	defer server.Close()
	handle(mux, `{"ok":true, "result":[]}`)

	bot, err := telegram.NewBot(telegram.Settings{
		URL:    server.URL,
		Token:  token,
		Client: httpClient,
		Poller: &telegram.LongPoller{Timeout: time.Second},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := pooling.New(config.Opts.PoolingSize())
	defer pool.Close()

	tg := &Telegram{ctx: ctx, bot: bot, pool: pool}
	time.AfterFunc(3*time.Second, func() {
		tg.Shutdown()
		cancel()
	})

	got := tg.Serve()
	expected := ErrServiceClosed
	if got != expected {
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
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	done := make(chan bool, 1)

	httpClient, mux, server := helper.MockServer()
	defer server.Close()
	handle(mux, getUpdatesJSON)

	tg, cancel, err := newTelegram(httpClient, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if tg.store == nil {
		t.Fatalf("Open storage failed: %v", err)
	}
	defer tg.store.Close()

	tg.bot.Poller = telegram.NewMiddlewarePoller(tg.bot.Poller, func(update *telegram.Update) bool {
		switch {
		// case update.Callback != nil:
		case update.Message != nil:
			if err := tg.process(update.Message); err != nil {
				t.Fatalf("process telegram message failed: %v", err)
			} else {
				done <- true
			}
		default:
			t.Log("Unhandle")
		}
		return true
	})

	go func() {
		tg.bot.Start()
	}()

	for {
		select {
		case <-done:
			tg.Shutdown()
			time.Sleep(time.Second)
			cancel()
			return
		case <-time.After(120 * time.Second):
			done <- true
		}
	}
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
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
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
        "from": {
          "id": -100000001,
          "is_bot": false,
          "first_name": "Somebody",
          "language_code": "en"
        },
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

	tg, cancel, err := newTelegram(httpClient, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if tg.store == nil {
		t.Fatalf("Open storage failed: %v", err)
	}
	defer tg.store.Close()

	tg.bot.Poller = telegram.NewMiddlewarePoller(tg.bot.Poller, func(update *telegram.Update) bool {
		switch {
		// case update.Callback != nil:
		case update.Message != nil:
			if err := tg.process(update.Message); err != nil {
				t.Fatalf("process telegram message failed: %v", err)
			} else {
				done <- true
			}
		default:
			t.Log("Unhandle")
		}
		return true
	})

	go func() {
		tg.bot.Start()
	}()

	for {
		select {
		case <-done:
			tg.Shutdown()
			time.Sleep(time.Second)
			cancel()
			return
		case <-time.After(120 * time.Second):
			done <- true
		}
	}
}
