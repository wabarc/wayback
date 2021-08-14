// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package template // import "github.com/wabarc/wayback/template"

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/wabarc/logger"
)

//go:embed views/*.html
var templateFiles embed.FS

//go:embed assets/image/*
var imageFiles embed.FS

//go:embed assets/js/*.js
var javascriptFiles embed.FS

// Static assets.
var (
	JavascriptBundleChecksums map[string]string
	JavascriptBundles         map[string][]byte
)

// Collect archived struct
type Collect struct {
	Slot string `json:"slot"`
	Src  string `json:"src"`
	Dst  string `json:"dst"`
}

// Collector represents a group of Collect.
type Collector []Collect

type funcMap struct {
	router *mux.Router
}

// Template handles the templating system.
type Template struct {
	templates map[string]*template.Template
	funcMap   *funcMap
}

// New returns a new template engine.
func New(router *mux.Router) *Template {
	return &Template{
		templates: make(map[string]*template.Template),
		funcMap:   &funcMap{router},
	}
}

// ParseTemplates parses template files embed into the application.
func (t *Template) ParseTemplates() error {
	entries, err := templateFiles.ReadDir("views")
	if err != nil {
		logger.Error("read views directory failed, %v", err)
		return err
	}

	var templateContents strings.Builder
	for _, entry := range entries {
		filename := entry.Name()
		fileData, err := templateFiles.ReadFile("views/" + filename)
		if err != nil {
			logger.Error("read views file %s failed, %v", err)
			return err
		}
		logger.Debug("parsing: %s", filename)

		templateContents.Write(fileData)
		t.templates[filename] = template.Must(template.New("web").Funcs(t.funcMap.wrap()).Parse(templateContents.String()))
	}

	return nil
}

// Render template with Collector
func (t *Template) Render(name string, data interface{}) ([]byte, bool) {
	logger.Info("render template: %s", name)

	name = strings.TrimSuffix(name, ".html") + ".html"
	tpl, ok := t.templates[name]
	if !ok {
		logger.Error("the template %s does not exists", name)
		return []byte{}, false
	}

	var b bytes.Buffer
	if err := tpl.Execute(&b, data); err != nil {
		logger.Error("execute template failed: %v", err)
		return []byte{}, false
	}

	return b.Bytes(), true
}

// LoadImageFile loads an embed image file.
func LoadImageFile(filename string) ([]byte, error) {
	return imageFiles.ReadFile(fmt.Sprintf(`assets/image/%s`, filename))
}

// GenerateJavascriptBundles creates JS bundles.
func GenerateJavascriptBundles() error {
	var bundles = map[string][]string{
		"index": {
			"assets/js/index.js",
		},
		"service-worker": {
			"assets/js/service-worker.js",
		},
	}

	JavascriptBundles = make(map[string][]byte)
	JavascriptBundleChecksums = make(map[string]string)

	for bundle, srcFiles := range bundles {
		var buffer bytes.Buffer

		for _, srcFile := range srcFiles {
			fileData, err := javascriptFiles.ReadFile(srcFile)
			if err != nil {
				return err
			}

			buffer.Write(fileData)
		}

		JavascriptBundles[bundle] = buffer.Bytes()
		JavascriptBundleChecksums[bundle] = fmt.Sprintf("%x", sha256.Sum256(buffer.Bytes()))
	}

	return nil
}

func (f *funcMap) wrap() template.FuncMap {
	return template.FuncMap{
		"route": func(name string, args ...interface{}) string {
			return Path(f.router, name, args...)
		},
	}
}

// Path returns the defined route based on given arguments.
func Path(router *mux.Router, name string, args ...interface{}) string {
	route := router.Get(name)
	if route == nil {
		logger.Error("route not found: %s", name)
		return ""
	}

	var pairs []string
	for _, arg := range args {
		switch param := arg.(type) {
		case string:
			pairs = append(pairs, param)
		default:
		}
	}

	result, err := route.URLPath(pairs...)
	if err != nil {
		logger.Error("parse URL path failed: %v", err)
		return ""
	}

	return result.String()
}
