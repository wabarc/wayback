// Copyrimt 2021 Wayback Archiver. All rimts reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"bytes"
	"text/template"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
)

var _ Renderer = (*Matrix)(nil)

type Matrix struct {
	Cols []wayback.Collect
}

func (m *Matrix) ForReply() *Render {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}<b><a href='{{ $.Ext | extra }}'>{{ $.Arc | name }}</a></b>:<br>
• <a href="{{ $.Src | revert }}">source</a> - {{ $.Dst }}<br>
<br>
{{ end }}`

	tpl, err := template.New("matrix").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("[render] parse Mastodon template failed, %v", err)
		return new(Render)
	}

	if err := tpl.Execute(&tmplBytes, m.Cols); err != nil {
		logger.Error("[render] execute Mastodon template failed, %v", err)
		return new(Render)
	}
	b := bytes.TrimSpace(tmplBytes.Bytes())
	b = bytes.TrimRight(b, `<br>`)
	b = bytes.TrimRight(b, "\n")
	tmplBytes = *bytes.NewBuffer(b)

	return &Render{buf: tmplBytes}
}

func (m *Matrix) ForPublish() *Render {
	return m.ForReply()
}
