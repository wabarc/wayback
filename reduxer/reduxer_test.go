// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"bufio"
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

const content = `<html>
<head>
    <title>Example Domain</title>
</head>

<body>
<div>
    <h1>Example Domain</h1>
    <p>This domain is for use in illustrative examples in documents. You may use this
    domain in literature without prior coordination or asking for permission.</p>
    <p><a href="https://www.iana.org/domains/example">More information...</a></p>
    <p><img src="/image.png"></p>
</div>
</body>
</html>
`

func genImage(height, width int) bytes.Buffer {
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Colors are defined by Red, Green, Blue, Alpha uint8 values.
	cyan := color.RGBA{100, 200, 200, 0xff}

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2: // upper left quadrant
				img.Set(x, y, cyan)
			case x >= width/2 && y >= height/2: // lower right quadrant
				img.Set(x, y, color.White)
			default:
				// Use zero value.
			}
		}
	}

	var b bytes.Buffer
	f := bufio.NewWriter(&b)
	png.Encode(f, img) // Encode as PNG.

	return b
}

func handleResponse(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(content))
	case "/image.png":
		buf := genImage(36, 36)
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(buf.Bytes())
	}
}

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

	helper.Unsetenv("WAYBACK_STORAGE_DIR")
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

	bundle, ok := res.Load(Src(inp.String()))
	if !ok {
		t.Fatal("Unexpected bundles")
	}
	art := bundle.Artifact()
	if art.Img.Local == "" || art.PDF.Local == "" || art.Raw.Local == "" {
		t.Fatal("Unexpected file path")
	}
}

func TestCreateDir(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "reduxer-")
	if err != nil {
		t.Fatalf(`Unexpected create temp dir: %v`, err)
	}
	defer os.RemoveAll(dir)

	helper.Unsetenv("WAYBACK_STORAGE_DIR")
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

func TestSingleFile(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "reduxer-")
	if err != nil {
		t.Fatalf(`Unexpected create temp dir: %v`, err)
	}
	defer os.RemoveAll(dir)

	_, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	exp := `<img src="data:image/png;base64,`
	if strings.Contains(content, exp) {
		t.Fatal(`unexpected sample html page`)
	}

	uri := server.URL
	filename := helper.RandString(5, "")
	ctx := context.WithValue(context.Background(), ctxBasenameKey, filename)
	got := singleFile(ctx, strings.NewReader(content), dir, uri)
	buf, _ := os.ReadFile(got)
	if !strings.Contains(string(buf), exp) {
		t.Fatal(`unexpected archive webpage as a single file`)
	}
}
