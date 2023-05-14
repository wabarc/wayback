// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package slack // import "github.com/wabarc/wayback/service/slack"

import (
	"context"
	"encoding/json"
	"fmt"
	// "io"
	"net/http"
	"net/url"
	"os"
	// "strings"
	"log"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	// "github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slacktest"
	"github.com/slack-go/slack/socketmode"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	// "github.com/wabarc/wayback/storage"
)

var (
	appToken     = "xapp-1-A028RLDJKDHU-123407010000001-adsfjcjdkxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	botToken     = "xoxp-230644354353357-22343241312434-2286583762287-dasfdklxjcvkjlsadasdfasdfasd2341"
	channel      = "CH0AU7DJ"
	authTestJSON = `{
    "ok": true,
    "url": "https://waybackarchiver.slack.com/",
    "team": "Wayback Archiver Workspace",
    "user": "wabarcbot",
    "team_id": "T12345678",
    "user_id": "W12345678"
}`
	connOpenJSON = `{
    "ok": true,
    "url": "%s"
}`
	eventHello = `{
  "type": "hello",
  "num_connections": 4,
  "debug_info": {
    "host": "applink-7fc4fdbb64-4x5xq",
    "build_number": 10,
    "approximate_connection_time": 18060
  },
  "connection_info": {
    "app_id": "A01K58AR4RF"
  }
}
`
)

func handle(mux *http.ServeMux, updatesJSON string) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// b, _ := io.ReadAll(r.Body)
		// var dat map[string]interface{}
		// if err := json.Unmarshal(b, &dat); err != nil {
		// 	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		// 	return
		// }

		switch r.URL.Path {
		case "/auth.test":
			fmt.Fprintln(w, authTestJSON)
		case "/apps.connections.open":
			fmt.Fprintln(w, updatesJSON)
		case "/ws":
			wsHandler(w, r)
		default:
			fmt.Println(r.URL.Path, r)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	})
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	slacktest.Websocket(func(c *websocket.Conn) {
		// serverAddr := r.Context().Value(slacktest.ServerBotHubNameContextKey).(string)
		// go handlePendingMessages(c, serverAddr)
		for {
			var (
				err   error
				m     json.RawMessage
				mtype string
			)

			if mtype, m, err = slacktest.RTMRespEventType(c); err != nil {
				if websocket.IsUnexpectedCloseError(err) {
					return
				}

				log.Printf("read error: %s", err.Error())
				continue
			}

			switch mtype {
			case "ping":
				if err = slacktest.RTMRespPong(c, m); err != nil {
					log.Println("ping error:", err)
				}
			default:
				// sts.postProcessMessage(string(m), serverAddr)
			}
		}
	})(w, r)
}

func TestServe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test in short mode.")
	}

	os.Setenv("WAYBACK_SLACK_APP_TOKEN", appToken)
	os.Setenv("WAYBACK_SLACK_BOT_TOKEN", botToken)
	os.Setenv("WAYBACK_SLACK_CHANNEL", channel)

	parser := config.NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	opts.EnableServices(config.ServiceSlack.String())

	s := slacktest.NewTestServer()
	_, mux, server := helper.MockServer()
	defer server.Close()

	uri, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	s.ServerAddr = uri.Host
	go s.Start()
	defer s.Stop()

	handle(mux, fmt.Sprintf(connOpenJSON, s.GetWSURL()))

	bot := slack.New(
		opts.SlackBotToken(),
		slack.OptionAPIURL(s.GetAPIURL()),
		// slack.OptionAPIURL(server.URL+"/"),
		// slack.OptionHTTPClient(httpClient),
		slack.OptionDebug(opts.HasDebugMode()),
		slack.OptionAppLevelToken(opts.SlackAppToken()),
	)
	if bot == nil {
		t.Fatal("create slack bot instance failed")
	}

	client := socketmode.New(
		bot,
		// socketmode.OptionDebug(opts.HasDebugMode()),
		// socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	cfg := []pooling.Option{
		pooling.Capacity(opts.PoolingSize()),
		pooling.Timeout(opts.WaybackTimeout()),
		pooling.MaxRetries(opts.WaybackMaxRetries()),
	}
	ctx, cancel := context.WithCancel(context.Background())
	pool := pooling.New(ctx, cfg...)
	go pool.Roll()
	defer pool.Close()

	sl := &Slack{ctx: ctx, bot: bot, opts: opts, pool: pool, client: client}
	time.AfterFunc(3*time.Second, func() {
		sl.Shutdown()
		cancel()
	})
	got := sl.Serve()
	expected := ErrServiceClosed
	if got != expected {
		t.Errorf("Unexpected serve slack got %v instead of %v", got, expected)
	}
	time.Sleep(time.Second)
}
