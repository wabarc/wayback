// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
)

var (
	apiKey = `foo`

	respGetIndex = fmt.Sprintf(`{
  "uid": "%s",
  "name": "%s",
  "createdAt": "2022-02-10T07:45:15.628261Z",
  "updatedAt": "2022-02-21T15:28:43.496574Z",
  "primaryKey": "id"
}`, indexing, indexing)

	respInvalidRequest = fmt.Sprintf(`{
  "message": "Index %s not found.",
  "code": "index_not_found",
  "type": "invalid_request",
  "link": "https://docs.meilisearch.com/errors#index_not_found"
}`, indexing)

	respCreateIndex = fmt.Sprintf(`{
  "uid": 0,
  "indexUid": "%s",
  "status": "enqueued",
  "type": "indexCreation",
  "enqueuedAt": "2021-08-12T10:00:00.000000Z"
}`, indexing)

	respPushDocument = fmt.Sprintf(`{
    "uid": 1,
    "indexUid": "%s",
    "status": "enqueued",
    "type": "documentAddition",
    "enqueuedAt": "2021-08-11T09:25:53.000000Z"
}`, indexing)

	simple = []wayback.Collect{
		{
			Arc: config.SLOT_IA,
			Dst: "https://web.archive.org/web/20211000000001/https://example.com/",
			Src: "https://example.com/",
			Ext: config.SLOT_IA,
		},
		{
			Arc: config.SLOT_IS,
			Dst: "http://archive.today/abcdE",
			Src: "https://example.com/",
			Ext: config.SLOT_IS,
		},
		{
			Arc: config.SLOT_IP,
			Dst: "https://ipfs.io/ipfs/QmTbDmpvQ3cPZG6TA5tnar4ZG6q9JMBYVmX2n3wypMQMtr",
			Src: "https://example.com/",
			Ext: config.SLOT_IP,
		},
		{
			Arc: config.SLOT_PH,
			Dst: "http://telegra.ph/title-01-01",
			Src: "https://example.com/",
			Ext: config.SLOT_PH,
		},
	}
)

func TestExistIndex(t *testing.T) {
	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != `/indexes/`+indexing {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, respInvalidRequest)
			return
		}
		if !strings.Contains(r.Header.Get(`Authorization`), apiKey) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, respInvalidRequest)
			return
		}
		fmt.Fprintf(w, respGetIndex)
	})

	m := NewMeili(server.URL, apiKey, indexing)
	err := m.existIndex()
	if err != nil {
		t.Fatalf(`unexpected check index: %v`, err)
	}
}

func TestCreateIndex(t *testing.T) {
	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != `/indexes` {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, respInvalidRequest)
			return
		}
		if !strings.Contains(r.Header.Get(`Authorization`), apiKey) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, respInvalidRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, respCreateIndex)
	})

	m := NewMeili(server.URL, apiKey, indexing)
	err := m.createIndex()
	if err != nil {
		t.Fatalf(`unexpected create index: %v`, err)
	}
}

func TestPushDocument(t *testing.T) {
	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != fmt.Sprintf(`/indexes/%s/documents`, indexing) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, respInvalidRequest)
			return
		}
		if !strings.Contains(r.Header.Get(`Authorization`), apiKey) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, respInvalidRequest)
			return
		}
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, respInvalidRequest)
			return
		}

		var docs []document
		if err := json.Unmarshal(buf, &docs); err != nil {
			return
		}
		if len(docs) != 1 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, respInvalidRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, respPushDocument)
	})

	m := NewMeili(server.URL, apiKey, indexing)
	err := m.push(simple)
	if err != nil {
		t.Fatalf(`unexpected push document: %v`, err)
	}
}
