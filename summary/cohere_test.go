// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/cohere-ai/cohere-go"
	"github.com/wabarc/helper"
)

var (
	apiKey            = os.Getenv("COHERE_APIKEY")
	summarized        = "This is a summary of the test input."
	summarizeResponse = []byte(fmt.Sprintf(`{
    "summary": "%s"
}`, summarized))

	handleFunc = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/summarize":
			w.Write(summarizeResponse)
		}
	}
)

func TestNewCohere(t *testing.T) {
	if apiKey == "" {
		t.Skip(`Must set env "COHERE_APIKEY"`)
	}

	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", handleFunc)

	tests := []struct {
		desc      string
		client    *http.Client
		key       string
		expectErr bool
		expectNil bool
	}{
		{
			desc:      "Valid inputs",
			client:    httpClient,
			key:       "valid_api_key",
			expectErr: false,
			expectNil: false,
		},
		{
			desc:      "Invalid API key",
			client:    httpClient,
			key:       apiKey,
			expectErr: true,
			expectNil: true,
		},
		{
			desc:      "Nil http.Client",
			client:    nil,
			key:       apiKey,
			expectErr: false,
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cohere, err := NewCohere(tt.client, tt.key)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if tt.expectNil && cohere != nil {
				t.Errorf("Expected nil value for Cohere instance")
			}
			if !tt.expectNil && cohere == nil {
				t.Errorf("Unexpected nil value for Cohere instance")
			}
		})
	}
}

func TestCohere_Summarize(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:        "Empty string",
			input:       "",
			expected:    "",
			expectedErr: fmt.Errorf("text not found"),
		},
		{
			name:        "Valid input",
			input:       "This is a test input for summarization.",
			expected:    summarized,
			expectedErr: nil,
		},
	}

	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", handleFunc)

	cohereClient := &cohere.Client{Client: *httpClient, BaseURL: server.URL + "/"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coh := &Cohere{client: cohereClient}

			// Call the Summarize method
			actual, actualErr := coh.Summarize(tt.input)

			// Check the results
			if tt.expected != actual {
				t.Fatalf(`unexpected summarize, got "%v" instead of "%v"`, actual, tt.expected)
			}
			if !reflect.DeepEqual(tt.expectedErr, actualErr) {
				t.Fatalf(`unexpected summarize, got "%v" instead of "%v"`, actualErr, tt.expectedErr)
			}
		})
	}
}
