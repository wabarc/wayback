// Copyrimt 2021 Wayback Archiver. All rimts reserved.
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

var _ Renderer = (*Matrix)(nil)

// Matrix represents a Matrix template data for render.
type Matrix struct {
	Cols []wayback.Collect
	Data interface{}
}

// ForReply implements the standard Renderer interface:
// it reads `[]wayback.Collect` from the Matrix and returns a *Render.
func (m *Matrix) ForReply() *Render {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}<b><a href='{{ $.Ext | extra }}'>{{ $.Arc | name }}</a></b>:<br>
• <a href="{{ $.Src | revert }}">source</a> - {{ $.Dst }}<br>
<br>
{{ end }}`

	tpl, err := template.New("matrix").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Mastodon template failed, %v", err)
		return new(Render)
	}

	if err := tpl.Execute(&tmplBytes, m.Cols); err != nil {
		logger.Error("execute Mastodon template failed, %v", err)
		return new(Render)
	}
	b := bytes.TrimSpace(tmplBytes.Bytes())
	b = bytes.TrimRight(b, `<br>`)
	b = bytes.TrimRight(b, "\n")
	tmplBytes = *bytes.NewBuffer(b)
	for _, bundle := range bundles(m.Data) {
		m.renderAssets(bundle.Assets, &tmplBytes)
	}

	return &Render{buf: tmplBytes}
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` and `reduxer.Bundle` from
// the Matrix and returns a *Render.
func (m *Matrix) ForPublish() *Render {
	var tmplBytes bytes.Buffer

	bundle := bundle(m.Data)
	if head := Title(bundle); head != "" {
		tmplBytes.WriteString(`‹ <b>`)
		tmplBytes.WriteString(head)
		tmplBytes.WriteString(`</b> ›<br><br>`)
	}
	if dgst := Digest(bundle); dgst != "" {
		tmplBytes.WriteString(dgst)
		tmplBytes.WriteString(`<br><br>`)
	}

	const tmpl = `{{range $ := .}}<b><a href='{{ $.Ext | extra }}'>{{ $.Arc | name }}</a></b>:<br>
• <a href="{{ $.Src | revert }}">source</a> - {{ $.Dst }}<br>
<br>
{{ end }}`

	tpl, err := template.New("matrix").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Mastodon template failed, %v", err)
		return new(Render)
	}

	if err := tpl.Execute(&tmplBytes, m.Cols); err != nil {
		logger.Error("execute Mastodon template failed, %v", err)
		return new(Render)
	}
	if bundle != nil {
		m.renderAssets(bundle.Assets, &tmplBytes)
	}
	b := bytes.TrimSpace(tmplBytes.Bytes())
	b = bytes.TrimRight(b, `<br>`)
	b = bytes.TrimRight(b, "\n")
	tmplBytes = *bytes.NewBuffer(b)

	return &Render{buf: tmplBytes}
}

func (m *Matrix) renderAssets(assets reduxer.Assets, tmplBytes *bytes.Buffer) {
	tmpl := `<b><a href="https://anonfiles.com/">AnonFiles</a></b> - [ <a href="{{ .Img.Remote.Anonfile -}}
">IMG</a> ¦ <a href="{{ .PDF.Remote.Anonfile }}">PDF</a> ¦ <a href="{{ .Raw.Remote.Anonfile -}}
">RAW</a> ¦ <a href="{{ .Txt.Remote.Anonfile }}">TXT</a> ¦ <a href="{{ .HAR.Remote.Anonfile -}}
">HAR</a> ¦ <a href="{{ .WARC.Remote.Anonfile }}">WARC</a> ¦ <a href="{{ .Media.Remote.Anonfile }}">MEDIA</a> ]<br>
<b><a href="https://catbox.moe/">Catbox</a></b> - [ <a href="{{ .Img.Remote.Catbox -}}
">IMG</a> ¦ <a href="{{ .PDF.Remote.Catbox }}">PDF</a> ¦ <a href="{{ .Raw.Remote.Catbox -}}
">RAW</a> ¦ <a href="{{ .Txt.Remote.Catbox }}">TXT</a> ¦ <a href="{{ .HAR.Remote.Catbox -}}
">HAR</a> ¦ <a href="{{ .WARC.Remote.Catbox }}">WARC</a> ¦ <a href="{{ .Media.Remote.Catbox }}">MEDIA</a> ]`

	tpl, err := template.New("assets").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Telegram template failed, %v", err)
	}
	if err = tpl.Execute(tmplBytes, assets); err != nil {
		logger.Error("execute Telegram template failed, %v", err)
	}
}
