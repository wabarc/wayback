// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package mastodon // import "github.com/wabarc/wayback/service/mastodon"

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/wabarc/wayback/config"
)

func TestProcess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer zoo" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if r.URL.Path == "/api/v1/conversations" {
			fmt.Fprintln(w, `[{"id": "1", "unread":true, "last_status" : {"content": "foo https://example.com/ bar"}}]`)
		} else {
			fmt.Fprintln(w, `{"access_token": "zoo"}`)
		}
		return
	}))
	defer ts.Close()

	os.Setenv("WAYBACK_MASTODON_SERVER", ts.URL)
	os.Setenv("WAYBACK_MASTODON_KEY", "foo")
	os.Setenv("WAYBACK_MASTODON_SECRET", "bar")
	os.Setenv("WAYBACK_MASTODON_TOKEN", "zoo")

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse enviroment variables or flags failed, error: %v", err)
	}

	m := New(config.Opts)
	ctx := context.Background()
	convs, err := m.client.GetConversations(ctx, nil)
	t.Logf("Conversations: %v", convs)
	if err != nil {
		t.Fatalf("Mastodon: Get conversations failure, err: %v", err)
	}
	if len(convs) != 1 {
		t.Fatalf("result should be 1: %d", len(convs))
	}

	for _, conv := range convs {
		if err = m.process(ctx, conv); err != nil {
			t.Fatalf("should not be fail: %v", err)
		}
	}
}
