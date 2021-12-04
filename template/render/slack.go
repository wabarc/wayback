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

var _ Renderer = (*Slack)(nil)

// Slack represents a Slack template data for render.
type Slack struct {
	Cols []wayback.Collect
	Data interface{}
}

// ForReply implements the standard Renderer interface:
// it reads `[]wayback.Collect` from the Slack and returns a *Render.
func (s *Slack) ForReply() (r *Render) {
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
	for _, bundle := range bundles(s.Data) {
		s.renderAssets(bundle.Assets, &tmplBytes)
	}
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))

	return &Render{buf: tmplBytes}
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` and `reduxer.Bundle` from
// the Slack and returns a *Render.
func (s *Slack) ForPublish() (r *Render) {
	var tmplBytes bytes.Buffer

	bundle := bundle(s.Data)
	if head := Title(bundle); head != "" {
		tmplBytes.WriteString(`‹ `)
		tmplBytes.WriteString(head)
		tmplBytes.WriteString(" ›\n\n")
	}
	if dgst := Digest(bundle); dgst != "" {
		tmplBytes.WriteString(dgst)
		tmplBytes.WriteString("\n\n")
	}

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
	if bundle != nil {
		s.renderAssets(bundle.Assets, &tmplBytes)
	}
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))

	return &Render{buf: tmplBytes}
}

func (s *Slack) renderAssets(assets reduxer.Assets, tmplBytes *bytes.Buffer) {
	tmpl := `<https://anonfiles.com/|AnonFiles> - [ <{{ .Img.Remote.Anonfile | url -}}
|IMG> ¦ <{{ .PDF.Remote.Anonfile | url }}|PDF> ¦ <{{ .Raw.Remote.Anonfile | url -}}
|RAW> ¦ <{{ .Txt.Remote.Anonfile | url }}|TXT> ¦ <{{ .HAR.Remote.Anonfile | url -}}
|HAR> ¦ <{{ .WARC.Remote.Anonfile | url }}|WARC> ¦ <{{ .Media.Remote.Anonfile | url }}|MEDIA> ]
<https://catbox.moe/|Catbox> - [ <{{ .Img.Remote.Catbox | url -}}
|IMG> ¦ <{{ .PDF.Remote.Catbox | url }}|PDF> ¦ <{{ .Raw.Remote.Catbox | url -}}
|RAW> ¦ <{{ .Txt.Remote.Catbox | url }}|TXT> ¦ <{{ .HAR.Remote.Catbox | url -}}
|HAR> ¦ <{{ .WARC.Remote.Catbox | url }}|WARC> ¦ <{{ .Media.Remote.Catbox | url }}|MEDIA> ]`

	tpl, err := template.New("assets").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Telegram template failed, %v", err)
	}
	tmplBytes.WriteString("\n")
	if err = tpl.Execute(tmplBytes, assets); err != nil {
		logger.Error("execute Telegram template failed, %v", err)
	}
	tmplBytes.WriteString("\n")
}
