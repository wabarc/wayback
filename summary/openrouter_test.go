// Copyright 2026 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

func TestNewOpenRouter(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	handleFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/chat/completions":
			w.Write(summarizeResponse)
		}
	}
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
			t.Setenv("WAYBACK_LLM_PROVIDER", "openrouter")
			t.Setenv("WAYBACK_LLM_APIKEY", tt.key)

			parser := config.NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf("Parse environment variables or flags failed, error: %v", err)
			}

			op := NewOpenRouter(tt.client, opts)
			if !tt.expectNil && op == nil {
				t.Errorf("Unexpected nil value for OpenRouter instance")
			}
		})
	}
}

func TestOpenRouterSummarize(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		mockStatus  int
		mockBody    string
		expected    string
		expectedErr string
	}{
		{
			name:        "Empty string",
			input:       "",
			expected:    "",
			expectedErr: "text not found",
		},
		{
			name:       "Valid input",
			input:      "This is a test input for summarization.",
			mockStatus: 200,
			mockBody: `{
				"messages":[
					{"role":"user","content":"This is the summary."}
				]
			}`,
			expected:    "This is the summary.",
			expectedErr: "",
		},
		{
			name:        "API error status",
			input:       "Non-empty",
			mockStatus:  500,
			mockBody:    `{"error":"server"}`,
			expected:    "",
			expectedErr: "openrouter api error: status 500",
		},
	}

	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	// Register handler at expected endpoint path used by the client.
	mux.HandleFunc("/api/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		// optional: assert method and headers
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Find matching test case by inspecting body or rely on sequential handling.
		// For simplicity, read body and decide response based on test inputs:
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		r.Body.Close()

		switch {
		case strings.Contains(req.Messages[1].Content, "This is a test input for summarization."):
			w.WriteHeader(200)
			w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"This is the summary."}}]}`))
		case strings.Contains(req.Messages[1].Content, "Non-empty"):
			w.WriteHeader(500)
			w.Write([]byte("server error"))
		default:
			// default success
			w.WriteHeader(200)
			w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WAYBACK_LLM_PROVIDER", "openrouter")
			t.Setenv("WAYBACK_LLM_APIKEY", "test-key")

			parser := config.NewParser()
			opts, err := parser.ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf("Parse environment variables or flags failed, error: %v", err)
			}

			op := NewOpenRouter(httpClient, opts)

			actual, actualErr := op.Summarize(tt.input)

			if tt.expectedErr != "" {
				if actualErr == nil {
					t.Fatalf("expected error %q, got nil", tt.expectedErr)
				}
				if actualErr.Error() != tt.expectedErr {
					t.Fatalf("unexpected error, got %q expected %q", actualErr.Error(), tt.expectedErr)
				}
				return
			}

			if actualErr != nil {
				t.Fatalf("unexpected error: %v", actualErr)
			}
			if actual != tt.expected {
				t.Fatalf(`unexpected summary, got "%v" instead of "%v"`, actual, tt.expected)
			}
		})
	}
}
