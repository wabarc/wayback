// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

import (
	"fmt"
	"strings"

	"github.com/didasy/tldr"
)

const maxCharacters = 128

// Interface guard
var _ Summarizer = (*Legacy)(nil)

// Legacy implements the Summarizer interface using the tldr.Bag package to
// perform local summarization.
type Legacy struct {
	*tldr.Bag
}

// NewLegacy creates a new instance of the Legacy struct with a new tldr.Bag instance.
func NewLegacy() *Legacy {
	return &Legacy{tldr.New()}
}

// Summarize generates a summary of the input text using legacy summarization.
// It returns the summary as a string and any error that occurred during summarization.
func (l *Legacy) Summarize(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("text not found")
	}

	l.MaxCharacters = maxCharacters
	res, err := l.Bag.Summarize(s, 1)
	if err != nil {
		return "", fmt.Errorf("summarize failed: %v", err)
	}

	if len(res) == 0 {
		return s, nil
	}

	return res[0], nil
}
