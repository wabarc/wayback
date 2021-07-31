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
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/reduxer"
)

const (
	maxTitleLen  = 256
	maxDigestLen = 500
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

func (r *Render) String() string {
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

func bundle(data interface{}) *reduxer.Bundle {
	if bundle, ok := data.(*reduxer.Bundle); ok {
		return bundle
	}
	return new(reduxer.Bundle)
}

func bundles(data interface{}) reduxer.Bundles {
	if bundles, ok := data.(reduxer.Bundles); ok {
		return bundles
	}
	return make(reduxer.Bundles)
}

func Title(bundle *reduxer.Bundle) string {
	if bundle == nil {
		return ""
	}
	logger.Debug("extract title from reduxer bundle title: %s", bundle.Title)

	t := []rune(bundle.Title)
	l := len(t)
	if l > maxTitleLen {
		t = t[:maxTitleLen]
	}

	return strings.TrimSpace(string(t))
}

func Digest(bundle *reduxer.Bundle) string {
	if bundle == nil {
		return ""
	}
	logger.Debug("generate digest from article content: %s", bundle.Article.TextContent)

	txt := []rune(bundle.Article.TextContent)
	l := len(txt)
	switch {
	case l == 0:
		return ""
	case l > maxDigestLen:
		txt = txt[:maxDigestLen]
		return string(txt) + ` ...`
	default:
		return string(txt)
	}
}
