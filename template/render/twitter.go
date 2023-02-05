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

var _ Renderer = (*Twitter)(nil)

// Twitter represents a Twitter template data for render.
type Twitter struct {
	Cols []wayback.Collect
	Data reduxer.Reduxer
}

// ForReply implements the standard Renderer interface:
// it reads `[]wayback.Collect` from the Twitter and returns a *Render.
func (t *Twitter) ForReply() *Render {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}{{ if not $.Arc "ph" }}
• {{ $.Arc | name }}
{{ range $map := $.Dst -}}
{{ range $src, $dst := $map -}}
> {{ $dst }}
{{end}}{{end}}{{end}}{{end}}`

	tpl, err := template.New("twitter").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Twitter template failed: %v", err)
		return new(Render)
	}

	groups := groupBySlot(t.Cols)
	logger.Debug("for reply twitter: %#v", groups)

	tmplBytes.WriteString(original(groups))
	if err := tpl.Execute(&tmplBytes, groups); err != nil {
		logger.Error("execute Twitter template failed: %v", err)
		return new(Render)
	}
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))
	tmplBytes.WriteString("\n\n#wayback #存档")

	return &Render{buf: tmplBytes}
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` and `reduxer.Reduxer` from
// the Twitter and returns a *Render.
//
// ForPublish generate tweet of given wayback collects in Twitter struct.
// It excluded telegra.ph, because this link has been identified by Twitter.
func (t *Twitter) ForPublish() *Render {
	var tmplBytes bytes.Buffer

	if title := Title(t.Cols, t.Data); title != "" {
		tmplBytes.WriteString(`‹ `)
		tmplBytes.WriteString(title)
		tmplBytes.WriteString(" ›\n\n")
	}

	const tmpl = `{{range $ := .}}{{ if not $.Arc "ph" }}
• {{ $.Arc | name }}
> {{ $.Dst }}
{{end}}{{end}}`

	tpl, err := template.New("twitter").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Twitter template failed: %v", err)
		return new(Render)
	}

	tmplBytes.WriteString(original(t.Cols))
	if err := tpl.Execute(&tmplBytes, t.Cols); err != nil {
		logger.Error("execute Twitter template failed: %v", err)
		return new(Render)
	}
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))
	tmplBytes.WriteString("\n\n#wayback #存档")

	return &Render{buf: tmplBytes}
}
