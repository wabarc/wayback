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
	Data reduxer.Reduxer
}

// ForReply implements the standard Renderer interface:
// it returns a Render from the ForPublish.
func (gh *GitHub) ForReply() *Render {
	return gh.ForPublish()
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` and `reduxer.Reduxer` from
// the GitHub and returns a *Render.
func (gh *GitHub) ForPublish() *Render {
	var tmplBytes bytes.Buffer

	if dgst := Digest(gh.Cols, gh.Data); dgst != "" {
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

	writeArtifact(gh.Cols, gh.Data, func(art reduxer.Artifact) {
		tmplBytes.WriteString("\n")
		gh.parseArtifact(art, &tmplBytes)
	})

	return &Render{buf: tmplBytes}
}

func (gh *GitHub) parseArtifact(assets reduxer.Artifact, tmplBytes *bytes.Buffer) {
	tmpl := `**[AnonFiles](https://anonfiles.com/)** - [ [IMG]({{ .Img.Remote.Anonfile | url -}}
) ¦ [PDF]({{ .PDF.Remote.Anonfile | url }}) ¦ [RAW]({{ .Raw.Remote.Anonfile | url -}}
) ¦ [TXT]({{ .Txt.Remote.Anonfile | url }}) ¦ [HAR]({{ .HAR.Remote.Anonfile | url -}}
) ¦ [HTM]({{ .HTM.Remote.Anonfile | url }}) ¦ [WARC]({{ .WARC.Remote.Anonfile | url -}}
) ¦ [MEDIA]({{ .Media.Remote.Anonfile | url }}) ]
**[Catbox](https://catbox.moe/)** - [ [IMG]({{ .Img.Remote.Catbox | url -}}
) ¦ [PDF]({{ .PDF.Remote.Catbox | url }}) ¦ [RAW]({{ .Raw.Remote.Catbox | url -}}
) ¦ [TXT]({{ .Txt.Remote.Catbox | url }}) ¦ [HAR]({{ .HAR.Remote.Catbox | url -}}
) ¦ [HTM]({{ .HTM.Remote.Catbox | url }}) ¦ [WARC]({{ .WARC.Remote.Catbox | url -}}
) ¦ [MEDIA]({{ .Media.Remote.Catbox | url }}) ]`

	tpl, err := template.New("assets").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Telegram template failed, %v", err)
	}
	tmplBytes.WriteString("\n")
	if err = tpl.Execute(tmplBytes, assets); err != nil {
		logger.Error("execute Telegram template failed, %v", err)
	}
}
