// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/template/render"
)

func setMastodonEnv() {
	os.Setenv("WAYBACK_MASTODON_KEY", "foo")
	os.Setenv("WAYBACK_MASTODON_SECRET", "bar")
	os.Setenv("WAYBACK_MASTODON_TOKEN", "zoo")

	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()
}

func TestToMastodon(t *testing.T) {
	setMastodonEnv()

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

	os.Setenv("WAYBACK_MASTODON_SERVER", server.URL)

	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()

	mstdn := NewMastodon(nil)
	txt := render.ForPublish(&render.Telegram{Cols: collects}).String()
	got := mstdn.ToMastodon(context.Background(), txt, "")
	if !got {
		t.Errorf("Unexpected publish toot got %t instead of %t", got, true)
	}
}
