// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

import (
	"strings"

	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/ingress"
)

// Summarizer is the interface that wraps the basic Summarize method.
//
// Summarize takes in a string of text and returns a summary.
type Summarizer interface {
	Summarize(s string) (string, error)
}

// NewSummary creates and returns a Summarizer based on the configured LLM provider.
// It inspects opts.LLMProvider() (case-insensitive) and constructs a provider-specific
// handler. It falls back to the legacy summarizer implementation.
// The returned Summarizer wraps the chosen handler.
func NewSummary(opts *config.Options) Summarizer {
	switch strings.ToLower(opts.LLMProvider()) {
	case "cohere":
		return NewCohere(ingress.Client(), opts)
	case "openrouter":
		return NewOpenRouter(ingress.Client(), opts)
	case "ollama":
		return NewOllama(ingress.Client(), opts)
	}

	return NewLegacy()
}
