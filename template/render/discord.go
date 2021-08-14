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

var _ Renderer = (*Discord)(nil)

// Discord represents a Discord template data for render.
type Discord struct {
	Cols []wayback.Collect
	Data interface{}
}

// ForReply implements the standard Renderer interface:
// it reads `[]wayback.Collect` from the Discord and returns a *Render.
func (d *Discord) ForReply() (r *Render) {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}{{ $.Arc | name }}:
• {{ $.Dst }}

{{end}}`

	tpl, err := template.New("message").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Discord template failed, %v", err)
		return r
	}

	if err = tpl.Execute(&tmplBytes, d.Cols); err != nil {
		logger.Error("execute Discord template failed, %v", err)
		return r
	}
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))

	return &Render{buf: tmplBytes}
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` and `reduxer.Bundle` from
// the Discord and returns a *Render.
func (d *Discord) ForPublish() (r *Render) {
	var tmplBytes bytes.Buffer

	if head := Title(bundle(d.Data)); head != "" {
		tmplBytes.WriteString(`**`)
		tmplBytes.WriteString(head)
		tmplBytes.WriteString(`**`)
		tmplBytes.WriteString("\n\n")
	}
	if dgst := Digest(bundle(d.Data)); dgst != "" {
		tmplBytes.WriteString(dgst)
		tmplBytes.WriteString("\n\n")
	}

	const tmpl = `{{range $ := .}}{{ $.Arc | name }}:
• {{ $.Dst }}

{{end}}`

	tpl, err := template.New("message").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Discord template failed, %v", err)
		return r
	}

	if err = tpl.Execute(&tmplBytes, d.Cols); err != nil {
		logger.Error("execute Discord template failed, %v", err)
		return r
	}
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))

	return &Render{buf: tmplBytes}
}
