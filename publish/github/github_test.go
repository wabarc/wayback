// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package github // import "github.com/wabarc/wayback/publish/github"

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"
)

func TestToIssues(t *testing.T) {
	t.Setenv("WAYBACK_GITHUB_TOKEN", "foo")
	t.Setenv("WAYBACK_GITHUB_OWNER", "bar")
	t.Setenv("WAYBACK_GITHUB_REPO", "zoo")
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/repos/bar/zoo/issues":
			body, _ := io.ReadAll(r.Body)
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

	gh := New(t.Context(), httpClient, opts)
	txt := render.ForPublish(&render.GitHub{Cols: publish.Collects, Data: reduxer.BundleExample()}).String()
	got := gh.toIssues(context.Background(), "", txt)
	if !got {
		t.Errorf("Unexpected create GitHub Issues got %t instead of %t", got, true)
	}
}

func TestShutdown(t *testing.T) {
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	httpClient, _, server := helper.MockServer()
	defer server.Close()

	gh := New(t.Context(), httpClient, opts)
	err := gh.Shutdown()
	if err != nil {
		t.Errorf("Unexpected shutdown: %v", err)
	}
}
