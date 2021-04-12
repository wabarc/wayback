// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

var toot = `Internet Archive:
• https://web.archive.org/web/20211000000001/https://example.com/?q=%E4%B8%AD%E6%96%87

archive.today:
• http://archive.today/abcdE

IPFS:
• https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr

Telegraph:
• http://telegra.ph/title-01-01
`

func setMastodonEnv() {
	os.Setenv("WAYBACK_MASTODON_KEY", "foo")
	os.Setenv("WAYBACK_MASTODON_SECRET", "bar")
	os.Setenv("WAYBACK_MASTODON_TOKEN", "zoo")
}

func TestRenderForMastodon(t *testing.T) {
	setMastodonEnv()

	mstdn := &Mastodon{}
	got := mstdn.Render(collects)
	if got != toot {
		t.Errorf("Unexpected render template for Mastodon got \n%s\ninstead of \n%s", got, toot)
	}
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
		r.ParseForm()
		switch r.URL.Path {
		case "/api/v1/statuses":
			status := r.FormValue("status")
			if status != toot {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"access_token": "zoo"}`)
		}
	})

	os.Setenv("WAYBACK_MASTODON_SERVER", server.URL)

	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()

	mstdn := NewMastodon(nil)
	got := mstdn.ToMastodon(context.Background(), mstdn.Render(collects), "")
	if !got {
		t.Errorf("Unexpected publish toot got %t instead of %t", got, true)
	}
}
