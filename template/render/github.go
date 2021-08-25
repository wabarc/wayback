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

var _ Renderer = (*GitHub)(nil)

// GitHub represents a GitHub template data for render.
type GitHub struct {
	Cols []wayback.Collect
	Data interface{}
}

// ForReply implements the standard Renderer interface:
// it returns a Render from the ForPublish.
func (gh *GitHub) ForReply() *Render {
	return gh.ForPublish()
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` and `reduxer.Bundle` from
// the GitHub and returns a *Render.
func (gh *GitHub) ForPublish() *Render {
	var tmplBytes bytes.Buffer

	bundle := bundle(gh.Data)
	if dgst := Digest(bundle); dgst != "" {
		tmplBytes.WriteString(dgst)
		tmplBytes.WriteString("\n\n")
	}

	const tmpl = `{{range $ := .}}**[{{ $.Arc | name }}]({{ $.Ext | extra }})**:
> source: [{{ $.Src | unescape | revert }}]({{ $.Src | revert }})
> archived: {{ if $.Dst | isURL }}[{{ $.Dst | unescape }}]({{ $.Dst | escapeString }})
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
	if bundle != nil {
		tmplBytes.WriteString("\n")
		gh.renderAssets(bundle.Assets, &tmplBytes)
	}

	return &Render{buf: tmplBytes}
}

func (gh *GitHub) renderAssets(assets reduxer.Assets, tmplBytes *bytes.Buffer) {
	tmpl := `**[AnonFiles](https://anonfiles.com/)** - [ [IMG]({{ .Img.Remote.Anonfile -}}
) ¦ [PDF]({{ .PDF.Remote.Anonfile }}) ¦ [RAW]({{ .Raw.Remote.Anonfile -}}
) ¦ [TXT]({{ .Txt.Remote.Anonfile }}) ¦ [HAR]({{ .HAR.Remote.Anonfile -}}
) ¦ [WARC]({{ .WARC.Remote.Anonfile }}) ¦ [MEDIA]({{ .Media.Remote.Anonfile }}) ]
**[Catbox](https://catbox.moe/)** - [ [IMG]({{ .Img.Remote.Catbox -}}
) ¦ [PDF]({{ .PDF.Remote.Catbox }}) ¦ [RAW]({{ .Raw.Remote.Catbox -}}
) ¦ [TXT]({{ .Txt.Remote.Catbox }}) ¦ [HAR]({{ .HAR.Remote.Catbox -}}
) ¦ [WARC]({{ .WARC.Remote.Catbox }}) ¦ [MEDIA]({{ .Media.Remote.Catbox }}) ]`

	tpl, err := template.New("assets").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Telegram template failed, %v", err)
	}
	tmplBytes.WriteString("\n")
	if err = tpl.Execute(tmplBytes, assets); err != nil {
		logger.Error("execute Telegram template failed, %v", err)
	}
}
