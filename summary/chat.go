// Copyright 2026 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

const systemPrompt = `You are a digital archivist and information synthesizer, your expertise lies in distilling "noise" from legacy web data into high-signal summaries.


Rules:
- Summary point must be anchored by specific verbatim quotes
- Ignore UI elements (navbars, footers) and focus on the core content
- Be objective, clinical, and precise. Strip away marketing fluff to reveal the underlying data
- Summary must be in the same language as the source content
- Do NOT repeat ideas from previous snapshots unless conditions have materially changed

The output should be a maximum of 280 plain text characters.`

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatContent struct {
	Type string
	Text string
}

type chatChoice struct {
	Message  chatMessage   `json:"message,omitempty"`
	Role     string        `json:"role"`
	Contents []chatContent `json:"content"`
}

type chatResponse struct {
	Message chatChoice   `json:"message,omitempty"`
	ID      string       `json:"id"`
	Choices []chatChoice `json:"choices,omitempty"`
}
