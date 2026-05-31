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
var _ Summarizer = (*Ollama)(nil)

// Ollama represents a text summarization client for Ollama LLM service.
type Ollama struct {
	client   *http.Client
	endpoint string
	model    string
}

// NewOllama creates a `Ollama` instance with the specified `http.Client` and options.
// If the `http.Client` instance is `nil`, the default client is used. This function returns a pointer
// to the newly created `Ollama` instance and an error, if any.
func NewOllama(c *http.Client, opts *config.Options) *Ollama {
	if c == nil {
		c = ingress.Client()
	}
	model := opts.LLMModel()
	if model == "" {
		model = "llama3.1:8b"
	}
	endpoint := strings.TrimSuffix(opts.LLMBaseURL(), "/")
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	return &Ollama{
		client:   c,
		endpoint: endpoint,
		model:    model,
	}
}

// Summarize generates a summary of the input text using Ollama's AI models.
// Returns the generated summary as a string and an error, if any.
func (o *Ollama) Summarize(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("text not found")
	}

	body := chatRequest{
		Model: o.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: s},
		},
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %v", err)
	}

	endpoint := o.endpoint + "/v1/chat/completions"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama api error: status %d", res.StatusCode)
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
