// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

/*
Package config handles configuration management for the application.
*/

package config // import "github.com/wabarc/wayback/config"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBoolValue(t *testing.T) {
	scenarios := map[string]bool{
		"":        true,
		"1":       true,
		"YES":     true,
		"Yes":     true,
		"yes":     true,
		"TRUE":    true,
		"True":    true,
		"true":    true,
		"on":      true,
		"false":   false,
		"off":     false,
		"invalid": false,
	}

	for input, expected := range scenarios {
		result := parseBool(input, true)
		if result != expected {
			t.Errorf(`Unexpected result for %q, got %v instead of %v`, input, result, expected)
		}
	}
}

func TestParseIntValue(t *testing.T) {
	if parseInt("2020", 1128) != 2020 {
		t.Errorf(`Defined variables should returns the specified value`)
	}
}

func TestParseIntValueWithUnsetVariable(t *testing.T) {
	if parseInt("", 1128) != 1128 {
		t.Errorf(`Unset variables should returns the default value`)
	}
}

func TestParseIntValueWithInvalidInput(t *testing.T) {
	if parseInt("invalid integer", 1128) != 1128 {
		t.Errorf(`Invalid integer should returns the default value`)
	}
}

func TestParseStringValue(t *testing.T) {
	if parseString("test", "default value") != "test" {
		t.Errorf(`Defined variables should returns the specified value`)
	}
}

func TestParseStringValueWithUnsetVariable(t *testing.T) {
	if parseString("", "default value") != "default value" {
		t.Errorf(`Unset variables should returns the default value`)
	}
}

func TestParseIntListValue(t *testing.T) {
	if len(parseIntList("2020,1128", []int{80})) != 2 {
		t.Errorf(`Defined variables should returns the specified value`)
	}
}

func TestGetDefaultFilenames(t *testing.T) {
	files := defaultFilenames()
	got := len(files)
	expected := 3
	if got != expected {
		t.Errorf(`Unexpected file path got %d instead of %d`, got, expected)
	}

	home, _ := os.UserHomeDir()
	paths := fmt.Sprintf("%s %s %s", "wayback.conf", filepath.Join(home, "wayback.conf"), "/etc/wayback.conf")
	for _, path := range files {
		if strings.Index(paths, path) < 0 {
			t.Errorf(`Unexpected file path got %s instead within '%s'`, path, paths)
		}
	}
}
