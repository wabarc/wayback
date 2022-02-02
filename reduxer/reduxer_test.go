// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

func TestDo(t *testing.T) {
	binPath := helper.FindChromeExecPath()
	if _, err := exec.LookPath(binPath); err != nil {
		t.Skip("Chrome headless browser no found, skipped")
	}

	dir, err := os.MkdirTemp(os.TempDir(), "reduxer-")
	if err != nil {
		t.Fatalf(`Unexpected create temp dir: %v`, err)
	}
	defer os.RemoveAll(dir)

	os.Clearenv()
	os.Setenv("WAYBACK_STORAGE_DIR", dir)

	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	inp, err := url.Parse("https://example.com/")
	if err != nil {
		t.Fatalf("Unexpected parse url: %v", err)
	}
	res, err := Do(context.Background(), inp)
	if err != nil {
		t.Fatalf("Unexpected execute do: %v", err)
	}

	if len(res) == 0 {
		t.Errorf("Unexpected got res as 0")
	}

	for _, r := range res {
		if r.Assets.Img.Local == "" || r.Assets.PDF.Local == "" || r.Assets.Raw.Local == "" {
			t.Fatal("Unexpected file path")
		}
	}
}

func TestCreateDir(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "reduxer-")
	if err != nil {
		t.Fatalf(`Unexpected create temp dir: %v`, err)
	}
	defer os.RemoveAll(dir)

	os.Clearenv()
	os.Setenv("WAYBACK_STORAGE_DIR", dir)

	parser := config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	dir, err = createDir(dir)
	if err != nil {
		t.Fatalf("Unexpected execute create dir: %v", err)
	}

	file, err := os.Create(filepath.Join(dir, "foo.bar"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
}
