// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package render // import "github.com/wabarc/wayback/template/render"

import (
	"bytes"
	"net/url"
	"sort"
	"strings"
	"text/template"

	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/reduxer"
	"golang.org/x/net/html"
)

const (
	maxTitleLen  = 256
	maxDigestLen = 500
)

// Render represents a Render result.
type Render struct {
	buf bytes.Buffer
}

// Renderer is the interface that wraps the ForReply and ForPublish method.
type Renderer interface {
	// ForReply render text for reply to user.
	ForReply() *Render

	// ForPublish render text for publish.
	ForPublish() *Render
}

// ForReply handles render template for replying to user, it
// returns a Render.
func ForReply(r Renderer) *Render {
	return r.ForReply()
}

// ForPublish handles render template for publishing, it
// returns a Render.
func ForPublish(r Renderer) *Render {
	return r.ForPublish()
}

// String returns a string from the Render.
func (r *Render) String() string {
	if r != nil {
		return strings.TrimSpace(r.buf.String())
	}
	return ""
}

func funcMap() template.FuncMap {
	cache := "https://webcache.googleusercontent.com/search?q=cache:"
	return template.FuncMap{
		"escapeString": html.EscapeString,
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
		"url": func(s string) string {
			if helper.IsURL(s) {
				return s
			}
			return ""
		},
	}
}

// Collect represents a render data collection.
// Arc is name of the archive service,
// Dst mapping the original URL and archived destination URL,
// Ext is extra descriptions.
type Collect struct {
	Arc, Ext, Src string

	Dst []map[string]string // wayback results
}

// Collects represents a set of Collect in a map, and its key is a URL string.
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

func deDepURI(cols []wayback.Collect) map[string]bool {
	uris := make(map[string]bool)
	for _, col := range cols {
		uris[col.Src] = true
	}
	return uris
}

// Title returns the title of the webpage. Its maximum length is defined by `maxTitleLen`.
func Title(cols []wayback.Collect, rdx reduxer.Reduxer) (title string) {
	if rdx == nil {
		return
	}

	for uri := range deDepURI(cols) {
		if bundle, ok := rdx.Load(reduxer.Src(uri)); ok {
			if shots := bundle.Shots(); shots != nil {
				text := shots.Title
				logger.Debug("extract title from reduxer bundle title: %s", text)
				t := []rune(text)
				l := len(t)
				if l > maxTitleLen {
					t = t[:maxTitleLen]
				}
				title += strings.TrimSpace(string(t))
			}
		}
	}

	return
}

// digest returns digest of the webpage content. Its maximum length is defined by `maxDigestLen`.
func digest(cols []wayback.Collect, rdx reduxer.Reduxer) (dgst string) {
	if rdx == nil {
		return
	}

	for uri := range deDepURI(cols) {
		if bundle, ok := rdx.Load(reduxer.Src(uri)); ok {
			if text := bundle.Article().TextContent; text != "" {
				logger.Debug("generate digest from article content: %s", text)
				t := []rune(text)
				l := len(t)
				switch {
				case l == 0:
					continue
				case l > maxDigestLen:
					t = t[:maxDigestLen]
					dgst += string(t) + ` ...`
				default:
					dgst += string(t)
				}
			}
		}
	}

	return
}

// summary returns summary of the webpage content. Its maximum length is defined by `maxDigestLen`.
func summary(cols []wayback.Collect, rdx reduxer.Reduxer) (dgst string) {
	if rdx == nil {
		return
	}

	for uri := range deDepURI(cols) {
		if bundle, ok := rdx.Load(reduxer.Src(uri)); ok {
			if text := bundle.Summary(); text != "" {
				logger.Debug("extracted summary from article content: %s", text)
				t := []rune(text)
				l := len(t)
				switch {
				case l == 0:
					continue
				case l > maxDigestLen:
					t = t[:maxDigestLen]
					dgst += string(t) + ` ...`
				default:
					dgst += string(t)
				}
			}
		}
	}

	return
}

func summaryOrDigest(cols []wayback.Collect, rdx reduxer.Reduxer) string {
	if sum := summary(cols, rdx); sum != "" {
		return sum
	}

	return digest(cols, rdx)
}

// writeArtifact writes archived artifact of the webpage.
func writeArtifact(cols []wayback.Collect, rdx reduxer.Reduxer, fn func(art reduxer.Artifact)) {
	if rdx == nil {
		return
	}

	for uri := range deDepURI(cols) {
		if bundle, ok := rdx.Load(reduxer.Src(uri)); ok {
			fn(bundle.Artifact())
		}
	}
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
				for src := range dst {
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

	ss := make([]kv, 0, len(sm))
	for k, v := range sm {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	var sb strings.Builder
	sb.WriteString("• source\n")
	for _, kv := range ss {
		sb.WriteString(`> `)
		sb.WriteString(kv.Key)
		sb.WriteString("\n")
	}
	sb.WriteString("\n————\n")

	return sb.String()
}
