// Copyright 2021 Wayback Archiver. All ritts reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"bytes"
	"text/template"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/reduxer"
)

var _ Renderer = (*Nostr)(nil)

// Nostr represents a Nostr template data for render.
type Nostr struct {
	Data reduxer.Reduxer
	Cols []wayback.Collect
}

// ForReply implements the standard Renderer interface:
// it reads `[]wayback.Collect` from the Nostr and returns a *Render.
func (n *Nostr) ForReply() *Render {
	return n.ForPublish()
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` and `reduxer.Reduxer` from
// the Nostr and returns a *Render.
//
// ForPublish generate tweet of given wayback collects in Nostr struct.
// It excluded telegra.ph, because this link has been identified by Nostr.
func (n *Nostr) ForPublish() *Render {
	var tmplBytes bytes.Buffer

	if title := Title(n.Cols, n.Data); title != "" {
		tmplBytes.WriteString(`‹ `)
		tmplBytes.WriteString(title)
		tmplBytes.WriteString(" ›\n\n")
	}

	const tmpl = `{{range $ := .}}
• {{ $.Arc | name }}
> {{ $.Dst }}
{{end}}`

	tpl, err := template.New("nostr").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Nostr template failed: %v", err)
		return new(Render)
	}

	tmplBytes.WriteString(original(n.Cols))
	if err := tpl.Execute(&tmplBytes, n.Cols); err != nil {
		logger.Error("execute Nostr template failed: %v", err)
		return new(Render)
	}

	return &Render{buf: tmplBytes}
}
