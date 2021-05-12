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

func init() {
	os.Setenv("WAYBACK_GITHUB_TOKEN", "foo")
	os.Setenv("WAYBACK_GITHUB_OWNER", "bar")
	os.Setenv("WAYBACK_GITHUB_REPO", "zoo")

	config.Opts, _ = config.NewParser().ParseEnvironmentVariables()
}

func TestRenderForGitHub(t *testing.T) {
	expected := `**[Internet Archive](https://web.archive.org/)**:
> origin: [https://example.com/?q=中文](https://example.com/?q=%E4%B8%AD%E6%96%87)
> archived: [https://web.archive.org/web/20211000000001/https://example.com/?q=中文](https://web.archive.org/web/20211000000001/https://example.com/?q=%E4%B8%AD%E6%96%87)

**[archive.today](https://archive.today/)**:
> origin: [https://example.com/](https://example.com/)
> archived: [http://archive.today/abcdE](http://archive.today/abcdE)

**[IPFS](https://ipfs.github.io/public-gateway-checker/)**:
> origin: [https://example.com/](https://example.com/)
> archived: [https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr](https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr)

**[Telegraph](https://telegra.ph/)**:
> origin: [https://example.com/](https://example.com/)
> archived: [http://telegra.ph/title-01-01](http://telegra.ph/title-01-01)
`

	gh := NewGitHub(&http.Client{})
	got := gh.Render(collects)
	if got != expected {
		t.Errorf("Unexpected render template for GitHub Issues got \n%s\ninstead of \n%s", got, expected)
	}
}

func TestRenderForGitHubFlawed(t *testing.T) {
	expected := `**[Internet Archive](https://web.archive.org/)**:
> origin: [https://example.com/?q=中文](https://example.com/?q=%E4%B8%AD%E6%96%87)
> archived: Get "https://web.archive.org/save/https://example.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)

**[archive.today](https://archive.today/)**:
> origin: [https://example.com/](https://example.com/)
> archived: [http://archive.today/abcdE](http://archive.today/abcdE)

**[IPFS](https://ipfs.github.io/public-gateway-checker/)**:
> origin: [https://example.com/](https://example.com/)
> archived: Archive failed.

**[Telegraph](https://telegra.ph/)**:
> origin: [https://example.com/404](https://example.com/404)
> archived: [https://web.archive.org/*/https://webcache.googleusercontent.com/search?q=cache:https://example.com/404](https://web.archive.org/*/https://webcache.googleusercontent.com/search?q=cache:https://example.com/404)
`

	gh := NewGitHub(&http.Client{})
	got := gh.Render(flawed)
	if got != expected {
		t.Errorf("Unexpected render template for GitHub Issues got \n%s\ninstead of \n%s", got, expected)
	}
}

func TestToIssues(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/repos/bar/zoo/issues":
			body, _ := ioutil.ReadAll(r.Body)
			if strings.Index(string(body), config.SlotName(config.SLOT_IA)) == -1 {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			w.Header().Set("Status", "201 Created")
			fmt.Fprintln(w, `{"id": 1}`)
		}
	})

	gh := NewGitHub(httpClient)
	got := gh.ToIssues(context.Background(), gh.Render(collects))
	if !got {
		t.Errorf("Unexpected create GitHub Issues got %t instead of %t", got, true)
	}
}
