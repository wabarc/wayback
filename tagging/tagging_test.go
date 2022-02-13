// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package tagging // import "github.com/wabarc/wayback/tagging"

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/go-shiori/go-readability"
	"github.com/wabarc/wayback/config"
)

var (
	htmlContent = `<!doctype html>
<html>
<head>
    <title>Example Domain</title>

    <meta charset="utf-8" />
    <meta http-equiv="Content-type" content="text/html; charset=utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style type="text/css">
    body {
        background-color: #f0f0f2;
        margin: 0;
        padding: 0;
        font-family: -apple-system, system-ui, BlinkMacSystemFont, "Segoe UI", "Open Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;
    }
    div {
        width: 600px;
        margin: 5em auto;
        padding: 2em;
        background-color: #fdfdff;
        border-radius: 0.5em;
        box-shadow: 2px 3px 7px 2px rgba(0,0,0,0.02);
    }
    a:link, a:visited {
        color: #38488f;
        text-decoration: none;
    }
    @media (max-width: 700px) {
        div {
            margin: 0 auto;
            width: auto;
        }
    }
    </style>
</head>

<body>
<div>
    <h1>Example Domain</h1>
    <p>This domain is for use in illustrative examples in documents. You may use this
    domain in literature without prior coordination or asking for permission.</p>
    <p><a href="https://www.iana.org/domains/example">More information...</a></p>
</div>
</body>
</html>`
	textContent = `Example Domain
This domain is for use in illustrative examples in documents. You may use this domain in literature without prior coordination or asking for permission.

More information...`
)

func TestRetrieve(t *testing.T) {
	t.Parallel()

	os.Clearenv()
	os.Setenv("WAYBACK_MAX_TAG_SIZE", strconv.Itoa(3))

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	var tests = []struct {
		name    string
		content string
		contain string
	}{
		{
			name:    "empty",
			content: "",
			contain: "",
		},
		{
			name:    "html",
			content: htmlContent,
			contain: "#coordination #literature #permission",
		},
		{
			name:    "text",
			content: textContent,
			contain: "#information #more #asking",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := strings.NewReader(test.content)
			art, err := readability.FromReader(r, nil)
			if err != nil {
				t.Fatalf("Unexpected create readability.Article: %v", err)
			}

			tags, err := Retrieve(context.Background(), art)
			if test.name != "empty" && err != nil {
				t.Fatalf("Unexpected retrieve tags: %v", err)
			}
			for _, tag := range tags {
				if !strings.Contains(test.contain, tag) {
					t.Errorf("Unexpected retrieved tag got %s instead within %v", tag, test.contain)
				}
			}
		})
	}
}
