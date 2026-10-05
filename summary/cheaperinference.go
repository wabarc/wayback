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
var _ Summarizer = (*CheaperInference)(nil)

// CheaperInference represents a text summarization client for Cheaper Inference LLM service.
type CheaperInference struct {
	client *http.Client
	apiKey string
	model  string
}

// NewCheaperInference creates a `CheaperInference` instance with the specified `http.Client` and options.
// If the `http.Client` instance is `nil`, the default client is used. This function returns a pointer
// to the newly created `CheaperInference` instance and an error, if any.
func NewCheaperInference(c *http.Client, opts *config.Options) *CheaperInference {
	if c == nil {
		c = ingress.Client()
	}
	model := opts.LLMModel()
	if model == "" {
		model = "gpt-5.4-mini"
	}

	return &CheaperInference{
		client: c,
		apiKey: opts.LLMApiKey(),
		model:  model,
	}
}

// Summarize generates a summary of the input text using Cheaper Inference's AI models.
// Returns the generated summary as a string and an error, if any.
func (ci *CheaperInference) Summarize(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("text not found")
	}

	body := chatRequest{
		Model: ci.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: s},
		},
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %v", err)
	}

	endpoint := "https://api.cheaperinference.com/v1/chat/completions"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ci.apiKey)

	res, err := ci.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("cheaperinference api error: status %d", res.StatusCode)
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
