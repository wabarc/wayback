// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

import (
	"testing"

	"github.com/cohere-ai/cohere-go"
	"github.com/wabarc/helper"
)

func TestSummarize(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", handleFunc)

	cohereClient := &cohere.Client{Client: *httpClient, BaseURL: server.URL + "/"}
	coh := &Cohere{client: cohereClient}

	tests := []struct {
		name       string
		handler    interface{}
		input      string
		wantErr    bool
		errMessage string
	}{
		{
			name:       "valid Cohere handler",
			handler:    coh,
			input:      "This is a test string.",
			wantErr:    false,
			errMessage: "",
		},
		{
			name:       "valid Locally handler",
			handler:    NewLocally(),
			input:      "This is a test string.",
			wantErr:    false,
			errMessage: "",
		},
		{
			name:       "invalid handler",
			handler:    "invalid-handler",
			input:      "This is a test string.",
			wantErr:    true,
			errMessage: "invalid handler",
		},
		{
			name:       "empty input",
			handler:    coh,
			input:      "",
			wantErr:    true,
			errMessage: "text not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sum := &Summary{Handler: tt.handler}

			_, err := sum.Summarize(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf(`Unexpected error status. Got "%v", but wanted error="%v"`, err, tt.wantErr)
			}

			if tt.wantErr && err.Error() != tt.errMessage {
				t.Fatalf(`Unexpected error message. Got "%v", but wanted "%v"`, err.Error(), tt.errMessage)
			}
		})
	}
}
