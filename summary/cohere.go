// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/ingress"
)

// Interface guard
var _ Summarizer = (*Cohere)(nil)

// Cohere represents a text summarization algorithm powered by Cohere's AI models.
type Cohere struct {
	client *http.Client
	apiKey string
	model  string
}

// NewCohere creates a `Cohere` instance with the specified `http.Client` instance and API key.
// If the `http.Client` instance is `nil`, the default client is used. This function returns a pointer
// to the newly created `Cohere` instance and an error, if any.
func NewCohere(c *http.Client, opts *config.Options) *Cohere {
	if c == nil {
		c = ingress.Client()
	}
	model := opts.LLMModel()
	if model == "" {
		model = "command-a-03-2025"
	}

	return &Cohere{
		client: c,
		apiKey: opts.LLMApiKey(),
		model:  model,
	}
}

// Summarize generates a summary of the input text using Cohere's AI models.
// Returns the generated summary as a string and an error, if any.
func (coh *Cohere) Summarize(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("text not found")
	}

	body := chatRequest{
		Model: coh.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: s},
		},
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %v", err)
	}

	endpoint := "https://api.cohere.ai/v2/chat"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+coh.apiKey)

	res, err := coh.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("cohere api error: status %d", res.StatusCode)
	}

	var cr chatResponse
	if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
		return "", fmt.Errorf("failed to decode body: %v", err)
	}

	if len(cr.Message.Contents) > 0 && strings.TrimSpace(cr.Message.Contents[0].Text) != "" {
		return strings.TrimSpace(cr.Message.Contents[0].Text), nil
	}

	return s, nil
}
