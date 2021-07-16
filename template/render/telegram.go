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

var _ Renderer = (*Telegram)(nil)

type Telegram struct {
	Cols []wayback.Collect
}

func (t *Telegram) ForReply() (r *Render) {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}<b><a href='{{ $.Ext | extra }}'>{{ $.Arc | name }}</a></b>:
{{ range $map := $.Dst -}}
{{ range $src, $dst := $map -}}
• <a href="{{ $src | revert }}">source</a> - {{ if $dst | isURL }}<a href="{{ $dst }}">{{ $dst }}</a>{{ else }}{{ $dst }}{{ end }}
{{ end }}{{ end }}
{{ end }}`

	tpl, err := template.New("message").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Telegram template failed, %v", err)
		return r
	}

	groups := groupBySlot(t.Cols)
	logger.Debug("for reply telegram: %#v", groups)
	if err = tpl.Execute(&tmplBytes, groups); err != nil {
		logger.Error("execute Telegram template failed, %v", err)
		return r
	}
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))
	tmplBytes.WriteString("\n\n#wayback #存档")

	return &Render{buf: tmplBytes}
}

func (t *Telegram) ForPublish() (r *Render) {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}
<b><a href='{{ $.Ext | extra }}'>{{ $.Arc | name }}</a></b>:
• <a href="{{ $.Src | revert }}">source</a> - {{ if $.Dst | isURL }}<a href="{{ $.Dst }}">{{ $.Dst }}</a>{{ else }}{{ $.Dst }}{{ end }}
{{ end }}`

	tpl, err := template.New("message").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Telegram template failed, %v", err)
		return r
	}

	if err = tpl.Execute(&tmplBytes, t.Cols); err != nil {
		logger.Error("execute Telegram template failed, %v", err)
		return r
	}
	tmplBytes.WriteString("\n#wayback #存档")
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))

	return &Render{buf: tmplBytes}
}
