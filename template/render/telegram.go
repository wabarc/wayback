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

var _ Renderer = (*Telegram)(nil)

// Telegram represents a Telegram template data for render.
type Telegram struct {
	Cols []wayback.Collect
	Data reduxer.Reduxer
}

// ForReply implements the standard Renderer interface:
// it reads `[]wayback.Collect` from the Telegram and returns a *Render.
func (t *Telegram) ForReply() (r *Render) {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}<b><a href="{{ $.Ext | extra }}">{{ $.Arc | name }}</a></b>:
{{ range $map := $.Dst -}}
{{ range $src, $dst := $map -}}
• <a href="{{ $src | revert }}">source</a> - {{ if $dst | isURL }}<a href="{{ $dst }}">{{ $dst }}</a>{{ else }}{{ $dst | escapeString }}{{ end }}
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
	tmplBytes.WriteString("\n")

	writeArtifact(t.Cols, t.Data, func(art reduxer.Artifact) {
		t.parseArtifact(art, &tmplBytes)
	})

	tmplBytes.WriteString("\n#wayback #存档")

	return &Render{buf: tmplBytes}
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` and `reduxer.Reduxer` from
// the Telegram and returns a *Render.
func (t *Telegram) ForPublish() (r *Render) {
	var tmplBytes bytes.Buffer

	if title := Title(t.Cols, t.Data); title != "" {
		tmplBytes.WriteString("<b>")
		tmplBytes.WriteString(title)
		tmplBytes.WriteString("</b>\n\n")
	}

	if dgst := Digest(t.Cols, t.Data); dgst != "" {
		tmplBytes.WriteString(dgst)
		tmplBytes.WriteString("\n\n")
	}

	tmpl := `{{range $ := .}}
<b><a href="{{ $.Ext | extra }}">{{ $.Arc | name }}</a></b>:
• <a href="{{ $.Src | revert }}">source</a> - {{ if $.Dst | isURL }}<a href="{{ $.Dst }}">{{ $.Dst }}</a>{{ else }}{{ $.Dst | escapeString }}{{ end }}
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

	writeArtifact(t.Cols, t.Data, func(art reduxer.Artifact) {
		t.parseArtifact(art, &tmplBytes)
	})

	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))
	tmplBytes.WriteString("\n\n#wayback #存档")

	return &Render{buf: tmplBytes}
}

func (t *Telegram) parseArtifact(assets reduxer.Artifact, tmplBytes *bytes.Buffer) {
	tmpl := `<b><a href="https://catbox.moe/">Catbox</a></b> - [ <a href="{{ .Img.Remote.Catbox | url -}}
">IMG</a> ¦ <a href="{{ .PDF.Remote.Catbox | url }}">PDF</a> ¦ <a href="{{ .Raw.Remote.Catbox | url -}}
">RAW</a> ¦ <a href="{{ .Txt.Remote.Catbox | url }}">TXT</a> ¦ <a href="{{ .HAR.Remote.Catbox | url -}}
">HAR</a> ¦ <a href="{{ .HTM.Remote.Catbox | url }}">HTM</a> ¦ <a href="{{ .WARC.Remote.Catbox | url -}}
">WARC</a> ¦ <a href="{{ .Media.Remote.Catbox | url }}">MEDIA</a> ]`

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
