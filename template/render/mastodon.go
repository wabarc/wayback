// Copyright 2021 Wayback Archiver. All rights reserved.
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

var _ Renderer = (*Mastodon)(nil)

// Mastodon represents a Mastodon template data for render.
type Mastodon struct {
	Cols []wayback.Collect
	Data reduxer.Reduxer
}

// ForReply implements the standard Renderer interface:
// it returns a Render from the ForPublish.
func (m *Mastodon) ForReply() *Render {
	return m.ForPublish()
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` and `reduxer.Reduxer` from
// the Mastodon and returns a *Render.
func (m *Mastodon) ForPublish() *Render {
	var tmplBytes bytes.Buffer

	if title := Title(m.Cols, m.Data); title != "" {
		tmplBytes.WriteString(`‹ `)
		tmplBytes.WriteString(title)
		tmplBytes.WriteString(" ›\n\n")
	}

	const tmpl = `{{range $ := .}}{{ $.Arc | name }}:
• {{ $.Dst }}

{{end}}`

	tpl, err := template.New("mastodon").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("[masatodon] parse Mastodon template failed, %v", err)
		return new(Render)
	}

	tmplBytes.WriteString(original(m.Cols))
	err = tpl.Execute(&tmplBytes, m.Cols)
	if err != nil {
		logger.Error("[masatodon] execute Mastodon template failed, %v", err)
		return new(Render)
	}
	tmplBytes.WriteString("#wayback" + createTags(m.Cols, m.Data))
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))

	return &Render{buf: tmplBytes}
}
