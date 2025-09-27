// Copyriit 2023 Wayback Archiver. All riits reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"bytes"
	"text/template"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
)

var _ Renderer = (*XMPP)(nil)

// XMPP represents a XMPP template data for render.
type XMPP struct {
	Data interface{}
	Cols []wayback.Collect
}

// ForReply implements the standard Renderer interface:
// it returns a Render from the ForPublish.
func (x *XMPP) ForReply() *Render {
	return x.ForPublish()
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` from the XMPP and returns a *Render.
func (x *XMPP) ForPublish() (r *Render) {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}{{ $.Arc | name }}:
{{ range $map := $.Dst -}}
{{ range $src, $dst := $map -}}
• {{ if $dst | isURL }}{{ $dst }}{{ else }}{{ $dst | escapeString }}{{ end }}
{{ end }}{{ end }}
{{ end }}`

	tpl, err := template.New("message").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Telegram template failed, %v", err)
		return r
	}

	groups := groupBySlot(x.Cols)
	logger.Debug("for reply telegram: %#v", groups)
	if err = tpl.Execute(&tmplBytes, groups); err != nil {
		logger.Error("execute Telegram template failed, %v", err)
		return r
	}
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))
	tmplBytes.WriteString("\n")

	return &Render{buf: tmplBytes}
}
