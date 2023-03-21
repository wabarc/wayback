// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package storage // import "github.com/wabarc/wayback/storage"

import (
	"os"
	"path"
	"testing"

	"github.com/wabarc/wayback/config"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"empty path", ""},
		{"exist path", "bolt.db"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := path.Join(t.TempDir(), test.path)
			if test.path == "" {
				file = test.path
			}

			opts, err := config.NewParser().ParseEnvironmentVariables()
			if err != nil {
				t.Fatalf("Parse environment variables or flags failed, error: %v", err)
			}
			defer os.Remove(opts.BoltPathname())

			s, err := Open(opts, file)
			if err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
			defer s.db.Close()

			if s == nil {
				t.Fatalf("Storage instance is nil")
			}
			if s.db == nil {
				t.Fatalf("bolt.DB instance is nil")
			}
		})
	}
}

func TestClose(t *testing.T) {
	file := path.Join(t.TempDir(), "bolt.db")
	opts := &config.Options{}
	s, err := Open(opts, file)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	err = s.Close()
	if err != nil {
		t.Fatalf("failed to close database: %v", err)
	}

	if s.db.String() != `DB<"">` {
		t.Fatalf("failed to close database: %s", s.db)
	}
}
