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

var _ Renderer = (*Slack)(nil)

type Slack struct {
	Cols []wayback.Collect
}

func (s *Slack) ForReply() (r *Render) {
	return s.ForPublish()
}

func (s *Slack) ForPublish() (r *Render) {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}{{ $.Arc | name }}:
• {{ $.Dst }}

{{end}}`

	tpl, err := template.New("message").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Slack template failed, %v", err)
		return r
	}

	if err = tpl.Execute(&tmplBytes, s.Cols); err != nil {
		logger.Error("execute Slack template failed, %v", err)
		return r
	}
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))

	return &Render{buf: tmplBytes}
}
