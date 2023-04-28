// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package mastodon // import "github.com/wabarc/wayback/publish/mastodon"

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/template/render"
)

func setMastodonEnv(t *testing.T) {
	t.Setenv("WAYBACK_MASTODON_KEY", "foo")
	t.Setenv("WAYBACK_MASTODON_SECRET", "bar")
	t.Setenv("WAYBACK_MASTODON_TOKEN", "zoo")
}

func TestToMastodon(t *testing.T) {
	setMastodonEnv(t)

	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer zoo" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		switch r.URL.Path {
		case "/api/v1/statuses":
			status := r.FormValue("status")
			if !strings.Contains(status, `title`) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"access_token": "zoo"}`)
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	})

	t.Setenv("WAYBACK_MASTODON_SERVER", server.URL)
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	mstdn := New(http.Client{}, opts)
	txt := render.ForPublish(&render.Telegram{Cols: publish.Collects}).String()
	got := mstdn.toMastodon(context.Background(), txt, "")
	if !got {
		t.Errorf("Unexpected publish toot got %t instead of %t", got, true)
	}
}

func TestShutdown(t *testing.T) {
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	httpClient, _, server := helper.MockServer()
	defer server.Close()

	mstdn := New(*httpClient, opts)
	err := mstdn.Shutdown()
	if err != nil {
		t.Errorf("Unexpected shutdown: %v", err)
	}
}
