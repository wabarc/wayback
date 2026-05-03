// Copyright 2026 Wayback Archiver. All rights reserved.
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
var _ Summarizer = (*OpenRouter)(nil)

// OpenRouter represents a text summarization client for OpenRouter LLM service.
type OpenRouter struct {
	client *http.Client
	apiKey string
	model  string
}

// NewOpenRouter creates a `OpenRouter` instance with the specified `http.Client` and options.
// If the `http.Client` instance is `nil`, the default client is used. This function returns a pointer
// to the newly created `OpenRouter` instance and an error, if any.
func NewOpenRouter(c *http.Client, opts *config.Options) *OpenRouter {
	if c == nil {
		c = ingress.Client()
	}
	model := opts.LLMModel()
	if model == "" {
		model = "openrouter/auto"
	}

	return &OpenRouter{
		client: c,
		apiKey: opts.LLMApiKey(),
		model:  model,
	}
}

// Summarize generates a summary of the input text using OpenRouter's AI models.
// Returns the generated summary as a string and an error, if any.
func (or *OpenRouter) Summarize(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("text not found")
	}

	body := chatRequest{
		Model: or.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: s},
		},
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %v", err)
	}

	endpoint := "https://openrouter.ai/api/v1/chat/completions"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+or.apiKey)

	res, err := or.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("openrouter api error: status %d", res.StatusCode)
	}

	var cr chatResponse
	if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
		return "", fmt.Errorf("failed to decode body: %v", err)
	}

	if len(cr.Choices) > 0 && strings.TrimSpace(cr.Choices[0].Message.Content) != "" {
		return strings.TrimSpace(cr.Choices[0].Message.Content), nil
	}

	return s, nil
}
