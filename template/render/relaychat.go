// Copyriit 2021 Wayback Archiver. All riits reserved.
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

var _ Renderer = (*Relaychat)(nil)

// Relaychat represents a Relaychat template data for render.
type Relaychat struct {
	Cols []wayback.Collect
	Data reduxer.Reduxer
}

// ForReply implements the standard Renderer interface:
// it returns a Render from the ForPublish.
func (i *Relaychat) ForReply() *Render {
	buf := i.join(i.main())

	return &Render{buf: *buf}
}

// ForPublish implements the standard Renderer interface:
// it reads `[]wayback.Collect` from the Relaychat and returns a *Render.
func (i *Relaychat) ForPublish() *Render {
	tmplBytes := &bytes.Buffer{}

	if title := Title(i.Cols, i.Data); title != "" {
		tmplBytes.WriteString(`‹ `)
		tmplBytes.WriteString(title)
		tmplBytes.WriteString(" ›")
		tmplBytes.WriteString("\n \n")
	}
	// tmplBytes.WriteString("Source:\n")
	tmplBytes.WriteString(original(i.Cols))
	tmplBytes.WriteString(" \n")
	tmplBytes.Write(i.main().Bytes())

	return &Render{buf: *i.join(tmplBytes)}
}

func (i *Relaychat) main() *bytes.Buffer {
	tmplBytes := new(bytes.Buffer)

	const tmpl = "{{range $ := .}}• {{ $.Arc | name }}:\n> {{ $.Dst }}\n{{end}}"

	tpl, err := template.New("relaychat").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse IRC template failed, %v", err)
		return new(bytes.Buffer)
	}

	if err := tpl.Execute(tmplBytes, i.Cols); err != nil {
		logger.Error("execute IRC template failed, %v", err)
		return new(bytes.Buffer)
	}

	return bytes.NewBuffer(bytes.TrimRight(tmplBytes.Bytes(), "\n"))
}

func (i *Relaychat) join(buf *bytes.Buffer) *bytes.Buffer {
	tmplBytes := &bytes.Buffer{}
	tmplBytes.WriteString("***** List of Archives *****")
	tmplBytes.WriteString("\n")
	tmplBytes.Write(buf.Bytes())
	tmplBytes.WriteString("\n")
	tmplBytes.WriteString("***** End of Archives *****")
	return tmplBytes
}
