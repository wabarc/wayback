// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package mastodon // import "github.com/wabarc/wayback/service/mastodon"

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
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
		if r.ParseForm() != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/conversations":
			fmt.Fprintln(w, `[{"id": "1", "unread":true, "last_status" : {"content": "foo https://example.com/ bar"}}]`)
		case "/api/v1/notifications":
			fmt.Fprintln(w, `[{"id": "1", "type": "mention", "status" : {"content": "foo https://example.com/ bar"}}]`)
		case "/api/v1/statuses":
			status := r.FormValue("status")
			if !strings.Contains(status, config.SlotName(config.SLOT_IA)) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{}`)
		case "/api/v1/notifications/dismiss":
			fmt.Fprintln(w, `{}`)
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
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	m := New(context.Background(), &storage.Storage{}, pooling.New(config.Opts.PoolingSize()))
	noti, err := m.client.GetNotifications(m.ctx, nil)
	if err != nil {
		t.Fatalf("Mastodon: Get notifications failure, err: %v", err)
	}
	if len(noti) != 1 {
		t.Fatalf("result should be 1: %d", len(noti))
	}

	for _, n := range noti {
		if err = m.process(n.ID, n.Status); err != nil {
			t.Fatalf("should not be fail: %v", err)
		}
	}
}

func TestPlayback(t *testing.T) {
	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer zoo" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if r.ParseForm() != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/conversations":
			fmt.Fprintln(w, `[{"id": "1", "unread":true, "last_status" : {"content": "foo /playback https://example.com/ bar"}}]`)
		case "/api/v1/notifications":
			fmt.Fprintln(w, `[{"id": "1", "type": "mention", "status" : {"content": "foo /playback https://example.com/ bar"}}]`)
		case "/api/v1/statuses":
			status := r.FormValue("status")
			if !strings.Contains(status, config.SlotName(config.SLOT_TT)) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{}`)
		case "/api/v1/notifications/dismiss":
			fmt.Fprintln(w, `{}`)
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
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	pool := pooling.New(config.Opts.PoolingSize())
	defer pool.Close()

	m := New(context.Background(), &storage.Storage{}, pool)
	noti, err := m.client.GetNotifications(m.ctx, nil)
	if err != nil {
		t.Fatalf("Mastodon: Get notifications failure, err: %v", err)
	}
	if len(noti) != 1 {
		t.Fatalf("result should be 1: %d", len(noti))
	}

	for _, n := range noti {
		if err = m.process(n.ID, n.Status); err != nil {
			t.Fatalf("should not be fail: %v", err)
		}
	}
}
