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

var _ Renderer = (*GitHub)(nil)

type GitHub struct {
	Cols []wayback.Collect
}

func (gh *GitHub) ForReply() *Render {
	return gh.ForPublish()
}

func (gh *GitHub) ForPublish() *Render {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}**[{{ $.Arc | name }}]({{ $.Ext | extra }})**:
> source: [{{ $.Src | unescape | revert }}]({{ $.Src | revert }})
> archived: {{ if $.Dst | isURL }}[{{ $.Dst | unescape }}]({{ $.Dst }})
{{ else }}{{ $.Dst }}
{{ end }}
{{ end }}`

	tpl, err := template.New("github").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse template failed, %v", err)
		return new(Render)
	}

	if err := tpl.Execute(&tmplBytes, gh.Cols); err != nil {
		logger.Error("execute template failed, %v", err)
		return new(Render)
	}
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))

	return &Render{buf: tmplBytes}
}
