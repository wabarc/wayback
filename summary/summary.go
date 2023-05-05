// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package summary // import "github.com/wabarc/wayback/summary"

// Summarizer is the interface that wraps the basic Summarize method.
//
// Summarize takes in a string of text and returns a summary.
type Summarizer interface {
	Summarize(s string) (string, error)
}
