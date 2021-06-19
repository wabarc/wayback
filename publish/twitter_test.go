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

	"github.com/dghubble/go-twitter/twitter"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

var tweet = `Origin:
• https://example.com/?q=%E4%B8%AD%E6%96%87

====

Internet Archive:
• https://web.archive.org/web/20211000000001/https://example.com/?q=%E4%B8%AD%E6%96%87

archive.today:
• http://archive.today/abcdE

IPFS:
• https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr`

func setTwitterEnv() {
	os.Setenv("WAYBACK_TWITTER_CONSUMER_KEY", "foo")
	os.Setenv("WAYBACK_TWITTER_CONSUMER_SECRET", "foo")
	os.Setenv("WAYBACK_TWITTER_ACCESS_TOKEN", "foo")
	os.Setenv("WAYBACK_TWITTER_ACCESS_SECRET", "foo")

	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()
}

func TestRenderForTwitter(t *testing.T) {
	setTwitterEnv()

	twitter := &Twitter{}
	got := twitter.Render(collects)
	if got != tweet {
		t.Errorf("Unexpected render template for Twitter got \n%s\ninstead of \n%s", got, tweet)
	}
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
			if status != tweet {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"id": 1}`)
		default:
			fmt.Fprintln(w, `{}`)
		}
	})

	twitt := NewTwitter(twitter.NewClient(httpClient))
	got := twitt.ToTwitter(context.Background(), twitt.Render(collects))
	if !got {
		t.Errorf("Unexpected create GitHub Issues got %t instead of %t", got, true)
	}
}
