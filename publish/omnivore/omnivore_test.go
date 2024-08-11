// Copyright 2024 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package omnivore // import "github.com/wabarc/wayback/publish/omnivore"

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
)

const saveURLResp = `{"data":{"saveUrl":{"url":"https://omnivore.app/repo/links/cff02ab5-c36e-4efe-a976-2de32dc1685d","clientRequestId":"cff02ab5-c36e-4efe-a976-2de32dc1685d"}}}`

func TestPublish(t *testing.T) {
	t.Setenv("WAYBACK_OMNIVORE_APIKEY", "foo")
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/graphql":
			fmt.Fprintln(w, saveURLResp)
		default:
			fmt.Fprintln(w, `{}`)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	o := New(httpClient, opts)
	got := o.Publish(ctx, reduxer.BundleExample(), publish.Collects)
	if got != nil {
		t.Errorf("unexpected save url got %v", got)
	}
}

func TestShutdown(t *testing.T) {
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	httpClient, _, server := helper.MockServer()
	defer server.Close()

	no := New(httpClient, opts)
	err := no.Shutdown()
	if err != nil {
		t.Errorf("Unexpected shutdown: %v", err)
	}
}
