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

	respGetVersion = `{
  "commitSha": "b46889b5f0f2f8b91438a08a358ba8f05fc09fc1",
  "commitDate": "2019-11-15T09:51:54.278247+00:00",
  "pkgVersion": "0.1.1"
}
`
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

	sample = []wayback.Collect{
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
	invalidSample = []wayback.Collect{
		{
			Arc: config.SLOT_IA,
			Dst: "invalid URL",
			Src: "https://example.com/",
			Ext: config.SLOT_IA,
		},
	}

	handlerFunc = func(w http.ResponseWriter, r *http.Request) {
		switch {
		case !strings.Contains(r.Header.Get(`Authorization`), apiKey):
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(respInvalidRequest))
		case r.URL.Path == `/version`:
			_, _ = w.Write([]byte(respGetVersion))
		case r.Method == http.MethodGet && r.URL.Path == `/indexes/`+indexing: // get index
			_, _ = w.Write([]byte(respGetIndex))
		case r.Method == http.MethodPost && r.URL.Path == `/indexes`: // create index
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(respCreateIndex))
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(`/indexes/%s/settings/sortable-attributes`, indexing): // set sortable attributes
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(respCreateIndex))
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(`/indexes/%s/documents`, indexing): // add documents
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(respInvalidRequest))
				return
			}

			var docs []document
			if err := json.Unmarshal(buf, &docs); err != nil {
				return
			}
			if len(docs) != 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(respInvalidRequest))
				return
			}

			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(respPushDocument))
		default:
			// Response invalid request
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(respInvalidRequest))
		}
	}
)

func TestNewMeili(t *testing.T) {
	tests := []struct {
		indexing string
		expected string
	}{
		{"", indexing},
		{"foo", "foo"},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			m := NewMeili("", "", test.indexing)
			if m.indexing != test.expected {
				t.Fatalf(`unexpected new meilisearch client got indexing name %s instead of %s`, m.indexing, test.expected)
			}
		})
	}
}

func TestSetup(t *testing.T) {
	tests := []struct {
		handler func(http.ResponseWriter, *http.Request)
		returns error
	}{
		{
			handler: handlerFunc,
			returns: nil,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			_, mux, server := helper.MockServer()
			defer server.Close()

			mux.HandleFunc("/", test.handler)

			m := NewMeili(server.URL, apiKey, indexing)
			err := m.Setup()
			if err != test.returns {
				t.Fatalf(`unexpected setup meilisearch, got error: %v`, err)
			}
		})
	}
}

func TestExistIndex(t *testing.T) {
	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", handlerFunc)

	m := NewMeili(server.URL, apiKey, indexing)
	err := m.existIndex()
	if err != nil {
		t.Fatalf(`unexpected check index: %v`, err)
	}
}

func TestCreateIndex(t *testing.T) {
	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", handlerFunc)

	m := NewMeili(server.URL, apiKey, indexing)
	err := m.createIndex()
	if err != nil {
		t.Fatalf(`unexpected create index: %v`, err)
	}
}

func TestPushDocument(t *testing.T) {
	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", handlerFunc)

	m := NewMeili(server.URL, apiKey, indexing)

	tests := []struct {
		collect []wayback.Collect
		returns string
	}{
		{
			collect: []wayback.Collect{},
			returns: `push documents failed: cols empty`,
		},
		{
			collect: sample,
			returns: ``,
		},
		{
			collect: invalidSample,
			returns: ``,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			err := m.push(test.collect)
			if err != nil && err.Error() != test.returns {
				t.Fatalf(`unexpected push document: %v`, err)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", handlerFunc)

	m := NewMeili(server.URL, apiKey, indexing)
	err := m.getVersion()
	if err != nil {
		t.Fatalf(`unexpected get version: %v`, err)
	}
	if m.version == "" {
		t.Fatal(`unexpected version`)
	}
}
