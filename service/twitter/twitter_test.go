// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package twitter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/dghubble/go-twitter/twitter"
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

// testServer returns an http Client, ServeMux, and Server. The client proxies
// requests to the server and handlers can be registered on the mux to handle
// requests. The caller must close the test server.
func testServer() (*http.Client, *http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	transport := &RewriteTransport{&http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}}
	client := &http.Client{Transport: transport}
	return client, mux, server
}

// RewriteTransport rewrites https requests to http to avoid TLS cert issues
// during testing.
type RewriteTransport struct {
	Transport http.RoundTripper
}

// RoundTrip rewrites the request scheme to http and calls through to the
// composed RoundTripper or if it is nil, to the http.DefaultTransport.
func (t *RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	if t.Transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.Transport.RoundTrip(req)
}

func TestProcess(t *testing.T) {
	httpClient, mux, server := testServer()
	defer server.Close()

	mux.HandleFunc("/1.1/direct_messages/events/new.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, testDMEventShowJSON)
	})
	mux.HandleFunc("/1.1/direct_messages/events/destroy.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(204)
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
