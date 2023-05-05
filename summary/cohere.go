// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cohere-ai/cohere-go"
)

// Interface guard
var _ Summarizer = (*Cohere)(nil)

// Cohere represents a text summarization algorithm powered by Cohere's AI models.
type Cohere struct {
	client *cohere.Client
}

// NewCohere creates a `Cohere` instance with the specified `http.Client` instance and API key.
// If the `http.Client` instance is `nil`, the default client is used. This function returns a pointer
// to the newly created `Cohere` instance and an error, if any.
func NewCohere(c *http.Client, key string) (*Cohere, error) {
	coh, err := cohere.CreateClient(key)
	if err != nil {
		return nil, err
	}
	if c != nil {
		coh.Client = *c
	}

	return &Cohere{coh}, nil
}

// Summarize generates a summary of the input text using Cohere's AI models.
// Returns the generated summary as a string and an error, if any.
func (coh *Cohere) Summarize(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("text not found")
	}

	res, err := coh.client.Summarize(cohere.SummarizeOptions{
		Text: s,
	})
	if err != nil {
		return "", err
	}

	return res.Summary, nil
}
