// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
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

func setMatrixEnv() {
	os.Setenv("WAYBACK_MATRIX_USERID", "@foo:example.com")
	os.Setenv("WAYBACK_MATRIX_ROOMID", "!bar:example.com")
	os.Setenv("WAYBACK_MATRIX_PASSWORD", "zoo")
}

func TestToMatrixRoom(t *testing.T) {
	unsetAllEnv()
	setMatrixEnv()

	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/_matrix/client/r0/login", r.URL.Path == "/_matrix/client/v3/login":
			fmt.Fprintln(w, `{"access_token": "zoo"}`)
		case strings.Contains(r.URL.Path, "!bar:example.com/send/m.room.message"):
			body, _ := ioutil.ReadAll(r.Body)
			if !strings.Contains(string(body), config.SlotName(config.SLOT_IA)) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, `{"id": 1}`)
		}
	})

	os.Setenv("WAYBACK_MATRIX_HOMESERVER", server.URL)
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	mat := NewMatrix(nil, opts)
	txt := render.ForPublish(&render.Mastodon{Cols: collects}).String()
	got := mat.toRoom(txt)
	if !got {
		t.Errorf("Unexpected publish room message got %t instead of %t", got, true)
	}
}
