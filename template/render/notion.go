// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"bytes"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/reduxer"
)

var _ Renderer = (*Notion)(nil)

// Notion represents a Notion template data for render.
type Notion struct {
	Cols []wayback.Collect
	Data reduxer.Reduxer
}

// ForReply implements the standard Renderer interface:
// it returns a Render from the ForPublish.
func (no *Notion) ForReply() *Render {
	return no.ForPublish()
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` and `reduxer.Reduxer` from
// the Notion and returns a *Render.
func (no *Notion) ForPublish() *Render {
	var tmplBytes bytes.Buffer

	rdx := no.Data
	for uri := range deDepURI(no.Cols) {
		if bundle, ok := rdx.Load(reduxer.Src(uri)); ok {
			if html := bundle.Article().Content; html != "" {
				logger.Debug("generate digest from article content: %s", html)
				tmplBytes.WriteString(html)
			}
		}
	}

	return &Render{buf: tmplBytes}
}
