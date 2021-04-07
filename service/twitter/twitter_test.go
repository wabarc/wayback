// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package twitter

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

// src: https://github.com/dghubble/go-twitter/blob/4b180d0cc78db653b2810d87f268590889f21a02/twitter/direct_messages_test.go#L12
var (
	testDMEvent = twitter.DirectMessageEvent{
		CreatedAt: "1542410751275",
		ID:        "1063573894173323269",
		Type:      "message_create",
		Message: &twitter.DirectMessageEventMessage{
			SenderID: "623265148",
			Target: &twitter.DirectMessageTarget{
				RecipientID: "3694959333",
			},
			Data: &twitter.DirectMessageData{
				Text: "foo https://example.com/ bar",
			},
		},
	}
	testDMEventJSON = `
{
	"type": "message_create",
	"id": "1063573894173323269",
	"created_timestamp": "1542410751275",
	"message_create": {
		"target": {
			"recipient_id": "3694959333"
		},
		"sender_id": "623265148",
		"message_data": {
			"text": "example",
			"entities": {
				"hashtags": [],
				"symbols": [],
				"user_mentions": [],
				"urls": []
			}
		}
  }
}`
	testDMEventShowJSON = `{"event": ` + testDMEventJSON + `}`
)

func TestProcess(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/1.1/direct_messages/events/new.json":
			fmt.Fprintf(w, testDMEventShowJSON)
		case "/1.1/direct_messages/events/destroy.json":
			w.WriteHeader(204)
		default:
			w.WriteHeader(404)
		}
	})

	os.Setenv("WAYBACK_TWITTER_CONSUMER_KEY", "foo")
	os.Setenv("WAYBACK_TWITTER_CONSUMER_SECRET", "foo")
	os.Setenv("WAYBACK_TWITTER_ACCESS_TOKEN", "foo")
	os.Setenv("WAYBACK_TWITTER_ACCESS_SECRET", "foo")

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse enviroment variables or flags failed, error: %v", err)
	}

	ctx := context.Background()
	client := twitter.NewClient(httpClient)
	tw := &Twitter{client: client, opts: config.Opts}
	if err := tw.process(ctx, testDMEvent); err != nil {
		t.Fatalf("should not be fail: %v", err)
	}
}
