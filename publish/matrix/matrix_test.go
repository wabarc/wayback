// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package matrix // import "github.com/wabarc/wayback/publish/matrix"

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/template/render"
)

func setMatrixEnv(t *testing.T) {
	t.Setenv("WAYBACK_MATRIX_USERID", "@foo:example.com")
	t.Setenv("WAYBACK_MATRIX_ROOMID", "!bar:example.com")
	t.Setenv("WAYBACK_MATRIX_PASSWORD", "zoo")
}

func matrixServer() *httptest.Server {
	_, mux, server := helper.MockServer()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/_matrix/client/r0/login", r.URL.Path == "/_matrix/client/v3/login":
			w.Write([]byte(`{"access_token": "zoo"}`))
		case strings.Contains(r.URL.Path, "!bar:example.com/send/m.room.message"):
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), config.SlotName(config.SLOT_IA)) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			w.Write([]byte(`{"id": 1}`))
		}
	})

	return server
}

func TestToMatrixRoom(t *testing.T) {
	setMatrixEnv(t)

	server := matrixServer()
	defer server.Close()

	t.Setenv("WAYBACK_MATRIX_HOMESERVER", server.URL)
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	mat := New(nil, opts)
	txt := render.ForPublish(&render.Mastodon{Cols: publish.Collects}).String()
	got := mat.toRoom(txt)
	if !got {
		t.Errorf("Unexpected publish room message got %t instead of %t", got, true)
	}
}

func TestShutdown(t *testing.T) {
	setMatrixEnv(t)

	server := matrixServer()
	defer server.Close()

	t.Setenv("WAYBACK_MATRIX_HOMESERVER", server.URL)
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	mat := New(nil, opts)
	err := mat.Shutdown()
	if err != nil {
		t.Errorf("Unexpected shutdown: %v", err)
	}
}
