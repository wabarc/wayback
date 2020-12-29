// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package template // import "github.com/wabarc/wayback/template"

import (
	"bytes"
	"github.com/wabarc/wayback/logger"
	"text/template"
)

// Collect archived struct
type Collect struct {
	Slot string `json:"slot"`
	Src  string `json:"src"`
	Dst  string `json:"dst"`
}

type Collector []Collect

var templates = template.Must(template.New("").Parse(html))

// Render template with Collector
func (c Collector) Render() ([]byte, bool) {
	var tpl bytes.Buffer
	if err := templates.Execute(&tpl, c); err != nil {
		logger.Error("Execute template failed, %v", err)
		return []byte{}, false
	} else {
		return tpl.Bytes(), true
	}
}
