// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

import (
	"fmt"
	"strings"

	"github.com/JesusIslam/tldr"
)

const maxCharacters = 128

// Interface guard
var _ Summarizer = (*Locally)(nil)

// Locally implements the Summarizer interface using the tldr.Bag package to
// perform local summarization.
type Locally struct {
	*tldr.Bag
}

// NewLocally creates a new instance of the Locally struct with a new tldr.Bag instance.
func NewLocally() *Locally {
	return &Locally{tldr.New()}
}

// Summarize generates a summary of the input text using local summarization.
// It returns the summary as a string and any error that occurred during summarization.
func (l *Locally) Summarize(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("text not found")
	}

	l.Bag.MaxCharacters = maxCharacters
	res, err := l.Bag.Summarize(s, 1)
	if err != nil {
		return "", fmt.Errorf("summarize failed: %v", err)
	}

	if len(res) == 0 {
		return s, nil
	}

	return res[0], nil
}
