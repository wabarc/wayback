// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"bytes"
	"net/url"
	"strings"
	"text/template"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
)

type Render struct {
	buf bytes.Buffer
}

type Renderer interface {
	ForReply() *Render
	ForPublish() *Render
}

func ForReply(r Renderer) *Render {
	return r.ForReply()
}

func ForPublish(r Renderer) *Render {
	return r.ForPublish()
}

func (r *Render) String() (text string) {
	if r != nil {
		return r.buf.String()
	}
	return ""
}

func funcMap() template.FuncMap {
	cache := "https://webcache.googleusercontent.com/search?q=cache:"
	return template.FuncMap{
		"unescape": func(link string) string {
			unescaped, err := url.QueryUnescape(link)
			if err != nil {
				return link
			}
			return unescaped
		},
		"isURL": helper.IsURL,
		"name":  config.SlotName,
		"extra": config.SlotExtra,
		"revert": func(link string) string {
			return strings.Replace(link, cache, "", 1)
		},
		"not": func(text, s string) bool {
			return !strings.Contains(text, s)
		},
	}
}

type Collect struct {
	Arc, Ext, Src string

	Dst []map[string]string
}

type Collects map[string]Collect

func groupBySlot(cols []wayback.Collect) *Collects {
	m := make(map[string][]map[string]string)
	for _, col := range cols {
		m[col.Arc] = append(m[col.Arc], map[string]string{col.Src: col.Dst})
	}
	c := make(Collects)
	for _, col := range cols {
		c[col.Arc] = Collect{
			Arc: col.Arc,
			Ext: col.Ext,
			Src: col.Src,
			Dst: m[col.Arc],
		}
	}
	return &c
}
