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

type Discord struct {
	Cols []wayback.Collect
}

func (d *Discord) ForReply() (r *Render) {
	return d.ForPublish()
}

func (d *Discord) ForPublish() (r *Render) {
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
