// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package mastodon // import "github.com/wabarc/wayback/service/mastodon"

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/storage"
)

func TestProcess(t *testing.T) {
	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer zoo" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/conversations":
			fmt.Fprintln(w, `[{"id": "1", "unread":true, "last_status" : {"content": "foo https://example.com/ bar"}}]`)
		case "/api/v1/statuses":
			fmt.Fprintln(w, `{"access_token": "zoo"}`)
		}
	})

	os.Setenv("WAYBACK_MASTODON_SERVER", server.URL)
	os.Setenv("WAYBACK_MASTODON_KEY", "foo")
	os.Setenv("WAYBACK_MASTODON_SECRET", "bar")
	os.Setenv("WAYBACK_MASTODON_TOKEN", "zoo")
	os.Setenv("WAYBACK_ENABLE_IA", "true")

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse enviroment variables or flags failed, error: %v", err)
	}

	m := New(&storage.Storage{})
	ctx := context.Background()
	convs, err := m.client.GetConversations(ctx, nil)
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
