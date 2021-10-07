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
	"os"
	// "strings"
	"log"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	// "github.com/gorilla/websocket"
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
	connOpenJSON = `{
    "ok": true,
    "url": "http://%s/"
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

		connOpenJSON = fmt.Sprintf(connOpenJSON, r.URL.Host)
		switch r.URL.Path {
		case "/auth.test":
			fmt.Fprintln(w, connOpenJSON)
		case "/apps.connections.open":
			fmt.Fprintln(w, connOpenJSON)
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

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	s := slacktest.NewTestServer()
	t.Log(s.GetAPIURL())
	go s.Start()

	_, mux, server := helper.MockServer()
	defer server.Close()
	handle(mux, `{"ok":true, "result":[]}`)
	// s.Handle("/apps.connections.open", slacktest.Websocket(func(conn *websocket.Conn) {
	// 	s.SendToWebsocket("dafd")
	// 	// if err := slacktest.RTMServerSendGoodbye(conn); err != nil {
	// 	// 	log.Println("failed to send goodbye", err)
	// 	// }
	// }))
	s.Handle("/apps.connections.open", wsHandler)

	bot := slack.New(
		config.Opts.SlackBotToken(),
		slack.OptionAPIURL(s.GetAPIURL()),
		// slack.OptionAPIURL(server.URL+"/"),
		// slack.OptionHTTPClient(httpClient),
		slack.OptionDebug(config.Opts.HasDebugMode()),
		slack.OptionAppLevelToken(config.Opts.SlackAppToken()),
	)
	if bot == nil {
		t.Fatal("create slack bot instance failed")
	}

	client := socketmode.New(
		bot,
		// socketmode.OptionDebug(config.Opts.HasDebugMode()),
		// socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(3*time.Second, func() {
		cancel()
	})

	pool := pooling.New(config.Opts.PoolingSize())
	defer pool.Close()

	sl := &Slack{ctx: ctx, bot: bot, pool: pool, client: client}
	got := sl.Serve()
	expected := "done"
	if got.Error() != expected {
		t.Errorf("Unexpected serve slack got %v instead of %v", got, expected)
	}
	time.Sleep(time.Second)
}
