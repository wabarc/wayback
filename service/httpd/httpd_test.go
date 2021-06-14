// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package httpd // import "github.com/wabarc/wayback/service/httpd"

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/reduxer"
)

func TestTransform(t *testing.T) {
	os.Setenv("WAYBACK_ENABLE_IA", "true")

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	text := "some text https://example.com"
	urls := helper.MatchURL(text)
	rbes := []reduxer.Bundle{}
	col, _ := wayback.Wayback(urls, &rbes)
	collector := transform(col)

	bytes, err := json.Marshal(collector)
	if err != nil {
		t.Fatalf("Decode json failed: %v", err)
	}

	if strings.Index(string(bytes), config.SlotName(config.SLOT_IA)) == 0 {
		t.Errorf("Unexpected transform archived result %s instead containing %s", string(bytes), config.SlotName(config.SLOT_IA))
	}
}

func TestProcessRespStatus(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	defer server.Close()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		newWeb().process(w, r)
	})

	var tests = []struct {
		status int
		method string
		data   string
	}{
		{
			method: http.MethodGet,
			status: http.StatusNotModified,
			data:   `{"text":"", "data-type":"json"}`,
		},
		{
			method: http.MethodPost,
			status: http.StatusNotModified,
			data:   `{"text":"foo bar", "data-type":"json"}`,
		},
	}

	for _, test := range tests {
		req, err := http.NewRequest(test.method, server.URL, strings.NewReader(test.data))
		if err != nil {
			t.Fatalf("Unexpected new request: %v", err)
		}

		req.Header.Add("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("Unexpected response: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != test.status {
			t.Fatalf("Unexpected response code got %d instead of %d", resp.StatusCode, test.status)
		}
	}
}

func TestProcessContentType(t *testing.T) {
	os.Setenv("WAYBACK_ENABLE_IA", "true")

	var err error
	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	web := newWeb()
	web.handle(pooling.New(config.Opts.PoolingSize()))
	httpClient, mux, server := helper.MockServer()
	defer server.Close()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		web.process(w, r)
	})

	var tests = []struct {
		status      int
		method      string
		contentType string
		data        string
	}{
		{
			method:      http.MethodPost,
			status:      http.StatusOK,
			contentType: "application/json",
			data:        `text=https%3A%2F%2Fexample.com&data-type=json`,
		},
		{
			method:      http.MethodPost,
			status:      http.StatusOK,
			contentType: "text/html; charset=utf-8",
			data:        `text=https%3A%2F%2Fexample.com`,
		},
	}

	for _, test := range tests {
		req, err := http.NewRequest(test.method, server.URL, strings.NewReader(test.data))
		if err != nil {
			t.Fatalf("Unexpected new request: %v", err)
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("Unexpected response: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != test.status {
			t.Fatalf("Unexpected response code got %d instead of %d", resp.StatusCode, test.status)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Unexpected read body: %v", err)
		}
		if strings.Index(string(body), config.SlotName(config.SLOT_IA)) == 0 {
			t.Fatalf("Unexpected wayback results got %s no containing %q", string(body), config.SlotName(config.SLOT_IA))
		}
	}
}
