// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"bytes"
	"text/template"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
)

var _ Renderer = (*Mastodon)(nil)

type Mastodon struct {
	Cols []wayback.Collect
}

func (m *Mastodon) ForReply() *Render {
	return m.ForPublish()
}

func (m *Mastodon) ForPublish() *Render {
	var tmplBytes bytes.Buffer

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
	tmplBytes.WriteString("#wayback #存档")
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))

	return &Render{buf: tmplBytes}
}
