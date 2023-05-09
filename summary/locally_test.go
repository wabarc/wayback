// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

import (
	"testing"
)

func TestLocally(t *testing.T) {
	// Define test cases as a slice of structs.
	tests := []struct {
		name       string
		input      string
		want       string
		wantErr    bool
		errMessage string
	}{
		{
			name:       "valid input",
			input:      "This is a test string.",
			want:       "This is a test string.",
			wantErr:    false,
			errMessage: "",
		},
		{
			name:       "empty input",
			input:      "",
			want:       "",
			wantErr:    true,
			errMessage: "text not found",
		},
	}

	local := NewLocally()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := local.Summarize(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf(`Unexpected error status. Got "%v", but wanted error="%v"`, err, tt.wantErr)
			}

			if tt.wantErr && err.Error() != tt.errMessage {
				t.Fatalf(`Unexpected error message. Got "%v", but wanted "%v"`, err.Error(), tt.errMessage)
			}

			if !tt.wantErr && got != tt.want {
				t.Fatalf(`Unexpected summary. Got "%v", but wanted "%v"`, got, tt.want)
			}
		})
	}
}
