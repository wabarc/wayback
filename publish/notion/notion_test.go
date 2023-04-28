// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package notion // import "github.com/wabarc/wayback/publish/notion"

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/wabarc/helper"
	"github.com/wabarc/imgbb"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
)

const (
	createPageResp = `{
  "object": "page",
  "id": "59833787-2cf9-4fdf-8782-e53db20768a5",
  "created_time": "2022-03-01T19:05:00.000Z",
  "last_edited_time": "2022-03-01T19:05:00.000Z",
  "created_by": {
    "object": "user",
    "id": "ee5f0f84-409a-440f-983a-a5315961c6e4"
  },
  "last_edited_by": {
    "object": "user",
    "id": "ee5f0f84-409a-440f-983a-a5315961c6e4"
  },
  "cover": {
    "type": "external",
    "external": {
      "url": "https://upload.wikimedia.org/wikipedia/commons/6/62/Tuscankale.jpg"
    }
  },
  "icon": {
    "type": "emoji",
    "emoji": "🥬"
  },
  "parent": {
    "type": "database_id",
    "database_id": "d9824bdc-8445-4327-be8b-5b47500af6ce"
  },
  "archived": false,
  "properties": {
    "Food group": {
      "id": "AHk",
      "type": "select",
      "select": {
        "id": "de8b67ad-44df-4758-a6cb-cc6c49fa8fe2",
        "name": "🥦 Vegetable",
        "color": "yellow"
      }
    },
    "Price": {
      "id": "BJXS",
      "type": "number",
      "number": 2.5
    },
    "+1": {
      "id": "Iowm",
      "type": "people",
      "people": []
    },
    "Recipes": {
      "id": "YfIu",
      "type": "relation",
      "relation": []
    },
    "Description": {
      "id": "_Tc_",
      "type": "rich_text",
      "rich_text": [
        {
          "type": "text",
          "text": {
            "content": "A dark green leafy vegetable",
            "link": null
          },
          "annotations": {
            "bold": false,
            "italic": false,
            "strikethrough": false,
            "underline": false,
            "code": false,
            "color": "default"
          },
          "plain_text": "A dark green leafy vegetable",
          "href": null
        }
      ]
    },
    "In stock": {
      "id": "605Bq3F",
      "type": "checkbox",
      "checkbox": false
    },
    "Name": {
      "id": "title",
      "type": "title",
      "title": [
        {
          "type": "text",
          "text": {
            "content": "Tuscan Kale",
            "link": null
          },
          "annotations": {
            "bold": false,
            "italic": false,
            "strikethrough": false,
            "underline": false,
            "code": false,
            "color": "default"
          },
          "plain_text": "Tuscan Kale",
          "href": null
        }
      ]
    }
  },
  "url": "https://www.notion.so/Tuscan-Kale-598337872cf94fdf8782e53db20768a5"
}`
	imgbbResponse = `{
  "id": "2ndCYJK",
  "title": "c1f64245afb2",
  "url": "https://github.githubassets.com/images/icons/emoji/unicode/1f30e.png",
  "image": {
    "filename": "1f30e.png",
    "name": "c1f64245afb2",
    "mime": "image/png",
    "extension": "png",
    "url": "https://github.githubassets.com/images/icons/emoji/unicode/1f30e.png"
  },
  "success": true,
  "status": 200
}`
	document = `<!DOCTYPE html>
<html>
  <head>
    <title>Example Domain</title>
  </head>

  <body>
    <div>
      <h1>Example Domain</h1>
      <p>This domain is for use in illustrative examples in documents. You may use this domain in literature without prior coordination or asking for permission.</p>
      <p><a href="https://www.iana.org/domains/example">More information...</a></p>
      <p><img alt="" src="https://example.com/images/dinosaur.jpg" /></p>
      <p>
        <video controls width="250">
          <source src="https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.webm" type="video/webm" />
          <source src="https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4" type="video/mp4" />
          Download the
          <a href="https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.webm">WEBM</a>
          or
          <a href="https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4">MP4</a>
          video.
        </video>
      </p>
      <p>
        <audio controls src="https://interactive-examples.mdn.mozilla.net/media/cc0-audio/t-rex-roar.mp3">
          <a href="https://interactive-examples.mdn.mozilla.net/media/cc0-audio/t-rex-roar.mp3">
            Download audio
          </a>
        </audio>
      </p>
      <p>
        <embed type="video/webm" src="https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4" width="250" height="200" />
      </p>
      <pre>
        <code>
			  L          TE
			    A       A
			      C    V
			       R A
			       DOU
			       LOU
			      REUSE
			      QUE TU
			      PORTES
			    ET QUI T'
			    ORNE O CI
			     VILISÉ
			    OTE-  TU VEUX
			     LA    BIEN
			    SI      RESPI
			            RER       - Apollinaire
        </code>
      </pre>
    </div>
  </body>
</html>`
)

func TestToNotion(t *testing.T) {
	t.Setenv("WAYBACK_NOTION_TOKEN", "foo")
	t.Setenv("WAYBACK_NOTION_DATABASE_ID", "bar")
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/pages":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "Example") {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, createPageResp)
		default:
			fmt.Fprintln(w, `{}`)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	no := New(httpClient, opts)
	got := no.Publish(ctx, reduxer.BundleExample(), publish.Collects)
	if got != nil {
		t.Errorf("unexpected create Notion got %v", got)
	}
}

func TestTraverseNodes(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/json":
			// Handles image upload to ImgBB
			fmt.Fprintln(w, imgbbResponse)
		default:
			fmt.Fprintln(w, `{}`)
		}
	})

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(document))
	if err != nil {
		t.Fatalf("unexpected new document: %v", err)
	}
	nodes := traverseNodes(doc.Contents(), imgbb.NewImgBB(httpClient, ""))
	if len(nodes) == 0 {
		t.Fatal("unexpected traverse nodes")
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
