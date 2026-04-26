// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

import (
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

func TestSummarize(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", handleFunc)

	t.Setenv("WAYBACK_LLM_PROVIDER", "cohere")
	t.Setenv("WAYBACK_LLM_APIKEY", "test-key")

	parser := config.NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	coh := NewCohere(httpClient, opts)

	tests := []struct {
		name       string
		handler    Summarizer
		input      string
		wantErr    bool
		errMessage string
	}{
		{
			name:       "Valid Cohere handler",
			handler:    coh,
			input:      "This is a test string.",
			wantErr:    false,
			errMessage: "",
		},
		{
			name:       "Valid Locally handler",
			handler:    NewLegacy(),
			input:      "This is a test string.",
			wantErr:    false,
			errMessage: "",
		},
		{
			name:       "Empty input",
			handler:    coh,
			input:      "",
			wantErr:    true,
			errMessage: "text not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.handler.Summarize(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf(`Unexpected error status. Got "%v", but wanted error="%v"`, err, tt.wantErr)
			}

			if tt.wantErr && err.Error() != tt.errMessage {
				t.Fatalf(`Unexpected error message. Got "%v", but wanted "%v"`, err.Error(), tt.errMessage)
			}
		})
	}
}
