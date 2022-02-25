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

	"github.com/dghubble/go-twitter/twitter"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/template/render"
)

func setTwitterEnv() {
	os.Setenv("WAYBACK_TWITTER_CONSUMER_KEY", "foo")
	os.Setenv("WAYBACK_TWITTER_CONSUMER_SECRET", "foo")
	os.Setenv("WAYBACK_TWITTER_ACCESS_TOKEN", "foo")
	os.Setenv("WAYBACK_TWITTER_ACCESS_SECRET", "foo")

	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()
}

func TestToTwitter(t *testing.T) {
	setTwitterEnv()

	httpClient, mux, server := helper.MockServer()
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

	twitt := NewTwitter(twitter.NewClient(httpClient))
	txt := render.ForPublish(&render.Twitter{Cols: collects}).String()
	got := twitt.ToTwitter(context.Background(), txt)
	if !got {
		t.Errorf("Unexpected create GitHub Issues got %t instead of %t", got, true)
	}
}
