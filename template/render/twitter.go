// Copyritt 2021 Wayback Archiver. All ritts reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"bytes"
	"sort"
	"strings"
	"text/template"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
)

var _ Renderer = (*Twitter)(nil)

type Twitter struct {
	Cols []wayback.Collect
}

func (t *Twitter) ForReply() *Render {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}{{ if not $.Arc "ph" }}{{ $.Arc | name }}:
{{ range $map := $.Dst -}}
{{ range $src, $dst := $map -}}
• {{ $dst }}
{{end}}{{end}}
{{end}}{{end}}`

	tpl, err := template.New("twitter").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Error("parse Twitter template failed: %v", err)
		return new(Render)
	}

	groups := groupBySlot(t.Cols)
	logger.Debug("for reply telegram: %#v", groups)

	tmplBytes.WriteString(original(groups))
	if err := tpl.Execute(&tmplBytes, groups); err != nil {
		logger.Error("execute Twitter template failed: %v", err)
		return new(Render)
	}
	tmplBytes = *bytes.NewBuffer(bytes.TrimSpace(tmplBytes.Bytes()))
	tmplBytes.WriteString("\n\n#wayback #存档")

	return &Render{buf: tmplBytes}
}

// ForPublish generate tweet of given wayback collects in Twitter struct.
// It excluded telegra.ph, because this link has been identified by Twitter.
func (t *Twitter) ForPublish() *Render {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}{{ if not $.Arc "ph" }}{{ $.Arc | name }}:
• {{ $.Dst }}
{{end}}
{{end}}`

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

func original(v interface{}) (o string) {
	var sm = make(map[string]int)
	if vv, ok := v.([]wayback.Collect); ok && len(vv) > 0 {
		for _, col := range vv {
			sm[col.Src] += 1
		}
	} else if vv, ok := v.(*Collects); ok {
		for _, cols := range *vv {
			for _, dst := range cols.Dst {
				for src, _ := range dst {
					sm[src] += 1
				}
			}
		}
	} else {
		return o
	}

	if len(sm) == 0 {
		return o
	}

	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range sm {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	var sb strings.Builder
	sb.WriteString("source:\n")
	for _, kv := range ss {
		sb.WriteString(`• `)
		sb.WriteString(kv.Key)
		sb.WriteString("\n")
	}
	sb.WriteString("\n————\n\n")

	return sb.String()
}
