// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package twitter // import "github.com/wabarc/wayback/publish/twitter"

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

func setTwitterEnv(t *testing.T) {
	t.Setenv("WAYBACK_TWITTER_CONSUMER_KEY", "foo")
	t.Setenv("WAYBACK_TWITTER_CONSUMER_SECRET", "foo")
	t.Setenv("WAYBACK_TWITTER_ACCESS_TOKEN", "foo")
	t.Setenv("WAYBACK_TWITTER_ACCESS_SECRET", "foo")
}

func TestToTwitter(t *testing.T) {
	setTwitterEnv(t)

	client, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/1.1/statuses/update.json":
			status := r.FormValue("status")
			if !strings.Contains(status, config.SlotName(config.SLOT_IA)) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"id": 1}`)
		default:
			fmt.Fprintln(w, `{}`)
		}
	})

	opts, _ := config.NewParser().ParseEnvironmentVariables()
	twitt := New(client, opts)
	txt := render.ForPublish(&render.Twitter{Cols: publish.Collects}).String()
	got := twitt.ToTwitter(context.Background(), txt)
	if !got {
		t.Errorf("Unexpected create GitHub Issues got %t instead of %t", got, true)
	}
}

func TestShutdown(t *testing.T) {
	setTwitterEnv(t)

	opts, _ := config.NewParser().ParseEnvironmentVariables()

	httpClient, _, server := helper.MockServer()
	defer server.Close()

	tw := New(httpClient, opts)
	err := tw.Shutdown()
	if err != nil {
		t.Errorf("Unexpected shutdown: %v", err)
	}
}
