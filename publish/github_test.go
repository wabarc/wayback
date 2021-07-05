// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/template/render"
)

func init() {
	os.Setenv("WAYBACK_GITHUB_TOKEN", "foo")
	os.Setenv("WAYBACK_GITHUB_OWNER", "bar")
	os.Setenv("WAYBACK_GITHUB_REPO", "zoo")

	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()
}

func TestToIssues(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/repos/bar/zoo/issues":
			body, _ := ioutil.ReadAll(r.Body)
			if !strings.Contains(string(body), config.SlotName(config.SLOT_IA)) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			w.Header().Set("Status", "201 Created")
			fmt.Fprintln(w, `{"id": 1}`)
		default:
			fmt.Fprintln(w, `{}`)
		}
	})

	gh := NewGitHub(httpClient)
	txt := render.ForPublish(&render.GitHub{Cols: collects}).String()
	got := gh.toIssues(context.Background(), nil, txt)
	if !got {
		t.Errorf("Unexpected create GitHub Issues got %t instead of %t", got, true)
	}
}
