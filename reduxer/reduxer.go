// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cixtor/readability"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/screenshot"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
)

type Path struct {
	Img, PDF, Raw string
}

type Bundle struct {
	screenshot.Screenshots

	Path    Path
	Article readability.Article
}

type Bundles map[string]Bundle

// Do executes secreenshot, print PDF and export html of given URLs
// Returns a set of bundle containing screenshot data and file path
func Do(ctx context.Context, urls ...string) (bundles Bundles, err error) {
	bundles = make(Bundles)
	if !config.Opts.EnabledReduxer() {
		return bundles, errors.New("Specify directory to environment `WAYBACK_STORAGE_DIR` to enable reduxer")
	}

	shots, err := Capture(ctx, urls...)
	if err != nil {
		return bundles, err
	}

	dir, err := createDir(config.Opts.StorageDir())
	if err != nil {
		return bundles, err
	}

	type m struct {
		key string
		val []byte
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var path Path
	for _, shot := range shots {
		wg.Add(1)
		go func(shot screenshot.Screenshots) {
			mu.Lock()
			defer mu.Unlock()
			defer wg.Done()
			slugs := []m{
				{key: "Img", val: shot.Image},
				{key: "PDF", val: shot.PDF},
				{key: "Raw", val: shot.HTML},
			}
			for _, slug := range slugs {
				if slug.val == nil {
					logger.Debug("[reduxer] file empty, skipped")
					continue
				}
				ft := http.DetectContentType(slug.val)
				fp := filepath.Join(dir, helper.FileName(shot.URL, ft))
				logger.Debug("[reduxer] writing file: %s", fp)
				if err := os.WriteFile(fp, slug.val, 0o600); err != nil {
					logger.Error("[reduxer] write %s file failed: %v", ft, err)
					continue
				}
				if err := helper.SetField(&path, slug.key, fp); err != nil {
					logger.Error("[reduxer] assign field to path struct failed: %v", err)
					continue
				}
			}
			bundle := Bundle{shot, path, readability.Article{}}
			article, err := readability.New().Parse(bytes.NewReader(shot.HTML), shot.URL)
			if err != nil {
				logger.Error("[reduxer] parse html failed: %v", err)
			}
			bundle.Article = article
			bundles[shot.URL] = bundle
		}(shot)
	}
	wg.Wait()

	return bundles, nil
}

// Capture returns screenshot.Screenshots of given URLs
func Capture(ctx context.Context, urls ...string) (shots []screenshot.Screenshots, err error) {
	opts := []screenshot.ScreenshotOption{
		screenshot.ScaleFactor(1),
		screenshot.PrintPDF(true), // print pdf
		screenshot.RawHTML(true),  // export html
		screenshot.Quality(100),   // image quality
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	shots = make([]screenshot.Screenshots, 0, len(urls))
	for _, uri := range urls {
		wg.Add(1)
		go func(uri string) {
			defer wg.Done()
			input, err := url.Parse(uri)
			if err != nil {
				logger.Error("[reduxer] parse url failed: %v", err)
				return
			}

			var shot screenshot.Screenshots
			if remote := remoteHeadless(config.Opts.ChromeRemoteAddr()); remote != nil {
				addr := remote.(*net.TCPAddr)
				headless, err := screenshot.NewChromeRemoteScreenshoter(addr.String())
				if err != nil {
					logger.Error("[reduxer] screenshot failed: %v", err)
					return
				}
				shot, err = headless.Screenshot(ctx, input, opts...)
			} else {
				shot, err = screenshot.Screenshot(ctx, input, opts...)
			}
			if err != nil {
				if err == context.DeadlineExceeded {
					logger.Error("[reduxer] screenshot deadline: %v", err)
					return
				}
				logger.Debug("[reduxer] screenshot error: %v", err)
				return
			}
			mu.Lock()
			shots = append(shots, shot)
			mu.Unlock()
		}(uri)
	}
	wg.Wait()

	return shots, nil
}

func remoteHeadless(addr string) net.Addr {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		logger.Debug("[reduxer] try to connect headless browser failed: %v", err)
		return nil
	}

	if conn != nil {
		conn.Close()
		logger.Debug("[reduxer] connected: %v", conn.RemoteAddr().String())
		return conn.RemoteAddr()
	} else {
		logger.Debug("[reduxer] headless chrome don't exists")
		return nil
	}
}

func createDir(baseDir string) (dir string, err error) {
	dir = filepath.Join(baseDir, time.Now().Format("200601"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		logger.Error("[reduxer] mkdir failed: %v", err)
		return "", err
	}
	return dir, nil
}
