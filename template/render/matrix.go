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
	Data reduxer.Reduxer
}

// ForReply implements the standard Renderer interface:
// it reads `[]wayback.Collect` from the Matrix and returns a *Render.
func (m *Matrix) ForReply() *Render {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}<b><a href='{{ $.Ext | extra }}'>{{ $.Arc | name }}</a></b>:<br>
• <a href="{{ $.Src | revert }}">source</a> - {{ $.Dst | escapeString }}<br>
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

	writeArtifact(m.Cols, m.Data, func(art reduxer.Artifact) {
		m.parseArtifact(art, &tmplBytes)
	})

	b := bytes.TrimSpace(tmplBytes.Bytes())
	b = bytes.TrimRight(b, `<br>`)
	b = bytes.TrimRight(b, "\n")
	tmplBytes = *bytes.NewBuffer(b)

	return &Render{buf: tmplBytes}
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` and `reduxer.Reduxer` from
// the Matrix and returns a *Render.
func (m *Matrix) ForPublish() *Render {
	var tmplBytes bytes.Buffer

	if title := Title(m.Cols, m.Data); title != "" {
		tmplBytes.WriteString(`‹ <b>`)
		tmplBytes.WriteString(title)
		tmplBytes.WriteString(`</b> ›<br><br>`)
	}

	if dgst := Digest(m.Cols, m.Data); dgst != "" {
		tmplBytes.WriteString(dgst)
		tmplBytes.WriteString(`<br><br>`)
	}

	const tmpl = `{{range $ := .}}<b><a href='{{ $.Ext | extra }}'>{{ $.Arc | name }}</a></b>:<br>
• <a href="{{ $.Src | revert }}">source</a> - {{ $.Dst | escapeString }}<br>
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

	writeArtifact(m.Cols, m.Data, func(art reduxer.Artifact) {
		m.parseArtifact(art, &tmplBytes)
	})

	b := bytes.TrimSpace(tmplBytes.Bytes())
	b = bytes.TrimRight(b, `<br>`)
	b = bytes.TrimRight(b, "\n")
	tmplBytes = *bytes.NewBuffer(b)

	return &Render{buf: tmplBytes}
}

func (m *Matrix) parseArtifact(assets reduxer.Artifact, tmplBytes *bytes.Buffer) {
	tmpl := `<b><a href="https://anonfiles.com/">AnonFiles</a></b> - [ <a href="{{ .Img.Remote.Anonfile | url -}}
">IMG</a> ¦ <a href="{{ .PDF.Remote.Anonfile | url }}">PDF</a> ¦ <a href="{{ .Raw.Remote.Anonfile | url -}}
">RAW</a> ¦ <a href="{{ .Txt.Remote.Anonfile | url }}">TXT</a> ¦ <a href="{{ .HAR.Remote.Anonfile | url -}}
">HAR</a> ¦ <a href="{{ .HTM.Remote.Anonfile | url }}">HTM</a> ¦ <a href="{{ .WARC.Remote.Anonfile | url -}}
">WARC</a> ¦ <a href="{{ .Media.Remote.Anonfile | url }}">MEDIA</a> ]<br>
<b><a href="https://catbox.moe/">Catbox</a></b> - [ <a href="{{ .Img.Remote.Catbox | url -}}
">IMG</a> ¦ <a href="{{ .PDF.Remote.Catbox | url }}">PDF</a> ¦ <a href="{{ .Raw.Remote.Catbox | url -}}
">RAW</a> ¦ <a href="{{ .Txt.Remote.Catbox | url }}">TXT</a> ¦ <a href="{{ .HAR.Remote.Catbox | url -}}
">HAR</a> ¦ <a href="{{ .HTM.Remote.Catbox | url }}">HTM</a> ¦ <a href="{{ .WARC.Remote.Catbox | url -}}
">WARC</a> ¦ <a href="{{ .Media.Remote.Catbox | url }}">MEDIA</a> ]`

	tpl, err := template.New("assets").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Telegram template failed, %v", err)
	}
	if err = tpl.Execute(tmplBytes, assets); err != nil {
		logger.Error("execute Telegram template failed, %v", err)
	}
}
