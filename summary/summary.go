// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

import (
	"fmt"
	"strings"
)

// Summarizer is the interface that wraps the basic Summarize method.
//
// Summarize takes in a string of text and returns a summary.
type Summarizer interface {
	Summarize(s string) (string, error)
}

// Interface guard
var _ Summarizer = (*Summary)(nil)

// Summary provides a high-level interface for generating text summaries using
// different summarization methods.
type Summary struct {
	Handler interface{}
}

// Summarize generates a summary of the input text using the selected summarization method.
// It returns the summary as a string and any error that occurred during summarization.
func (sum *Summary) Summarize(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("text not found")
	}

	switch handler := sum.Handler.(type) {
	case *Cohere:
		return handler.Summarize(s)
	case *Locally:
		return handler.Summarize(s)
	default:
		return "", fmt.Errorf("invalid handler")
	}
}
