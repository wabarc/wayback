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
)

var matExp = `<b><a href='https://web.archive.org/'>Internet Archive</a></b>:<br>
• <a href="https://example.com/?q=%E4%B8%AD%E6%96%87">source</a> - https://web.archive.org/web/20211000000001/https://example.com/?q=%E4%B8%AD%E6%96%87<br>
<br>
<b><a href='https://archive.today/'>archive.today</a></b>:<br>
• <a href="https://example.com/">source</a> - http://archive.today/abcdE<br>
<br>
<b><a href='https://ipfs.github.io/public-gateway-checker/'>IPFS</a></b>:<br>
• <a href="https://example.com/">source</a> - https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr<br>
<br>
<b><a href='https://telegra.ph/'>Telegraph</a></b>:<br>
• <a href="https://example.com/">source</a> - http://telegra.ph/title-01-01<br>
`

func setMatrixEnv() {
	os.Setenv("WAYBACK_MATRIX_USERID", "@foo:example.com")
	os.Setenv("WAYBACK_MATRIX_ROOMID", "!bar:example.com")
	os.Setenv("WAYBACK_MATRIX_PASSWORD", "zoo")
}

func TestRenderForMatrix(t *testing.T) {
	setMatrixEnv()

	mat := &Matrix{}
	got := mat.Render(collects)
	if got != matExp {
		t.Errorf("Unexpected render template for Matrix got \n%s\ninstead of \n%s", got, matExp)
	}
}

func TestToMatrixRoom(t *testing.T) {
	setMatrixEnv()

	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/_matrix/client/r0/login":
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
	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()

	mat := NewMatrix(nil)
	got := mat.ToRoom(context.Background(), mat.Render(collects))
	if !got {
		t.Errorf("Unexpected publish room message got %t instead of %t", got, true)
	}
}
