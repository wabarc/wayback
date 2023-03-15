// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-shiori/go-readability"
	"github.com/go-shiori/obelisk"
	"github.com/wabarc/go-anonfile"
	"github.com/wabarc/go-catbox"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/screenshot"
	"github.com/wabarc/warcraft"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"golang.org/x/sync/errgroup"
)

var (
	ctxBasenameKey struct{}

	filePerm = os.FileMode(0o600)
)

// Reduxer is the interface that wraps the basic reduxer method.
//
// Store sets the *bundle for a Src.
//
// Load returns the data stored in the map for a Src, or nil if no value is
// present. The ok result indicates whether value was found in the map.
//
// Flush erases all bundles from the cache.
type Reduxer interface {
	Store(Src, *bundle)
	Load(Src) (*bundle, bool)
	Flush()
}

// bundle represents a bundle data of a webpage.
type bundle struct {
	artifact Artifact
	article  readability.Article
	shots    *screenshot.Screenshots[screenshot.Path]
}

// Artifact represents the file paths stored on the local disk.
type Artifact struct {
	Img, PDF, Raw, Txt, HAR, HTM, WARC, Media Asset
}

// Asset represents the files on the local disk and the remote servers.
type Asset struct {
	Remote Remote
	Local  string
}

// Remote represents the file on the remote server.
type Remote struct {
	Anonfile string
	Catbox   string
}

// Src represents the requested url.
type Src string

// bundles represents a set of the bundle in a map, and its key is a URL string.
type bundles struct {
	mutex sync.RWMutex
	dirty map[Src]*bundle
}

// NewReduxer returns a Reduxer has been initialized.
func NewReduxer() Reduxer {
	return &bundles{
		mutex: sync.RWMutex{},
		dirty: make(map[Src]*bundle),
	}
}

// Store sets the *bundle for a Src.
func (bs *bundles) Store(key Src, b *bundle) {
	bs.mutex.Lock()
	if bs.dirty == nil {
		bs.dirty = make(map[Src]*bundle)
	}
	bs.dirty[key] = b
	bs.mutex.Unlock()
}

// Load returns the data stored in the map for a Src, or nil if no value is
// present. The ok result indicates whether value was found in the map.
func (bs *bundles) Load(key Src) (v *bundle, ok bool) {
	bs.mutex.RLock()
	v, ok = bs.dirty[key]
	bs.mutex.RUnlock()
	return
}

// Flush removes all bundles from the cache.
func (bs *bundles) Flush() {
	for key := range bs.dirty {
		bs.mutex.Lock()
		delete(bs.dirty, key)
		bs.mutex.Unlock()
	}
}

// Shots returns a screenshot.Screenshots from bundle.
func (b *bundle) Shots() *screenshot.Screenshots[screenshot.Path] {
	return b.shots
}

// Artifact returns an Artifact from bundle.
func (b *bundle) Artifact() Artifact {
	return b.artifact
}

// Article returns a readability.Article from bundle.
func (b *bundle) Article() readability.Article {
	return b.article
}

// Do executes secreenshot, print PDF and export html of given URLs
// Returns a set of bundle containing screenshot data and file path
// nolint:gocyclo
func Do(ctx context.Context, opts *config.Options, urls ...*url.URL) (Reduxer, error) {
	// Returns an initialized Reduxer for safe.
	var bs = NewReduxer()
	var err error

	if !opts.EnabledReduxer() {
		return bs, errors.New("Specify directory to environment `WAYBACK_STORAGE_DIR` to enable reduxer")
	}

	dir, err := createDir(opts.StorageDir())
	if err != nil {
		return bs, errors.Wrap(err, "create storage directory failed")
	}

	var warc = &warcraft.Warcraft{BasePath: dir, UserAgent: opts.WaybackUserAgent()}
	var craft = func(in *url.URL) (path string) {
		path, err = warc.Download(ctx, in)
		if err != nil {
			logger.Debug("create warc for %s failed: %v", in.String(), err)
			return ""
		}
		return path
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, uri := range urls {
		uri := uri
		g.Go(func() error {
			basename := strings.TrimSuffix(helper.FileName(uri.String(), ""), ".html")
			basename = strings.TrimSuffix(basename, ".htm")
			ctx = context.WithValue(ctx, ctxBasenameKey, basename)

			shot, er := capture(ctx, opts, uri, dir)
			if er != nil {
				return errors.Wrap(er, "capture failed")
			}
			logger.Debug("capture results: %#v", shot)

			artifact := &Artifact{
				Img:  Asset{Local: fmt.Sprint(shot.Image)},
				Raw:  Asset{Local: fmt.Sprint(shot.HTML)},
				PDF:  Asset{Local: fmt.Sprint(shot.PDF)},
				HAR:  Asset{Local: fmt.Sprint(shot.HAR)},
				WARC: Asset{Local: craft(uri)},
			}

			fp := filepath.Join(dir, basename)
			m := media{
				dir:  dir,
				path: fp,
				name: basename,
				url:  shot.URL,
			}

			if supportedMediaSite(uri) {
				artifact.Media.Local = m.download(ctx, opts)
			}
			// Attach single file
			var buf []byte
			var article readability.Article
			buf, err = os.ReadFile(fmt.Sprint(shot.HTML))
			if err == nil {
				singleFilePath := singleFile(ctx, bytes.NewReader(buf), dir, shot.URL)
				artifact.HTM.Local = singleFilePath
			}
			article, err = readability.FromReader(bytes.NewReader(buf), uri)
			if err != nil {
				logger.Error("parse html failed: %v", err)
			}
			txtName := basename + ".txt"
			fp = filepath.Join(dir, txtName)
			if err = os.WriteFile(fp, helper.String2Byte(article.TextContent), filePerm); err == nil && article.TextContent != "" {
				artifact.Txt.Local = fp
			}
			// Upload files to third-party server
			if err = remotely(ctx, artifact); err != nil {
				logger.Error("upload files to remote server failed: %v", err)
			}
			bundle := &bundle{shots: shot, artifact: *artifact, article: article}
			bs.Store(Src(shot.URL), bundle)
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return bs, errors.Wrap(err, "reduxer failed")
	}

	return bs, err
}

// capture returns screenshot.Screenshots of given URLs
func capture(ctx context.Context, cfg *config.Options, uri *url.URL, dir string) (shot *screenshot.Screenshots[screenshot.Path], err error) {
	filename := basename(ctx)
	files := screenshot.Files{
		Image: filepath.Join(dir, filename+".png"),
		HTML:  filepath.Join(dir, filename+".html"),
		PDF:   filepath.Join(dir, filename+".pdf"),
		HAR:   filepath.Join(dir, filename+".har"),
	}
	opts := []screenshot.ScreenshotOption{
		screenshot.AppendToFile(files),
		screenshot.ScaleFactor(1),
		screenshot.PrintPDF(true), // print pdf
		screenshot.DumpHAR(true),  // export har
		screenshot.RawHTML(true),  // export html
		screenshot.Quality(100),   // image quality
	}

	if remote := remoteHeadless(cfg.ChromeRemoteAddr()); remote != nil {
		logger.Debug("reduxer using remote browser")
		addr := remote.(*net.TCPAddr)
		browser, er := screenshot.NewChromeRemoteScreenshoter[screenshot.Path](addr.String())
		if er != nil {
			return shot, errors.Wrap(er, "dial screenshoter failed")
		}
		shot, err = browser.Screenshot(ctx, uri, opts...)
	} else {
		logger.Debug("reduxer using local browser")
		shot, err = screenshot.Screenshot[screenshot.Path](ctx, uri, opts...)
	}
	if err != nil {
		if err == context.DeadlineExceeded {
			return shot, errors.Wrap(err, "screenshot deadline")
		}
		return shot, errors.Wrap(err, "screenshot error")
	}

	return shot, err
}

func remoteHeadless(addr string) net.Addr {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return nil
	}

	if conn != nil {
		conn.Close()
		return conn.RemoteAddr()
	}
	return nil
}

func createDir(baseDir string) (dir string, err error) {
	dir = filepath.Join(baseDir, time.Now().Format("200601"))
	if helper.Exists(dir) {
		return
	}
	// nosemgrep
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", errors.Wrap(err, "mkdir failed: "+dir)
	}
	return dir, nil
}

func remotely(ctx context.Context, artifact *Artifact) (err error) {
	v := []*Asset{
		&artifact.Img,
		&artifact.PDF,
		&artifact.Raw,
		&artifact.Txt,
		&artifact.HAR,
		&artifact.HTM,
		&artifact.WARC,
		&artifact.Media,
	}

	c := &http.Client{}
	cat := catbox.New(c)
	anon := anonfile.NewAnonfile(c)
	g, _ := errgroup.WithContext(ctx)
	var mu sync.Mutex
	for _, asset := range v {
		asset := asset
		g.Go(func() error {
			mu.Lock()
			defer mu.Unlock()

			if asset.Local == "" {
				return nil
			}
			if !helper.Exists(asset.Local) {
				logger.Debug("local asset: %s not exists", asset.Local)
				return nil
			}
			r, e := anon.Upload(asset.Local)
			if e != nil {
				err = errors.Wrap(e, fmt.Sprintf("upload %s to anonfiles failed", asset.Local))
			} else {
				asset.Remote.Anonfile = r.Short()
			}
			c, e := cat.Upload(asset.Local)
			if e != nil {
				err = errors.Wrap(e, fmt.Sprintf("upload %s to catbox failed", asset.Local))
			} else {
				asset.Remote.Catbox = c
			}
			return err
		})
	}
	if err = g.Wait(); err != nil {
		return err
	}

	return nil
}

func singleFile(ctx context.Context, inp io.Reader, dir, uri string) string {
	req := obelisk.Request{URL: uri, Input: inp}
	arc := &obelisk.Archiver{
		SkipResourceURLError: true,
		RequestTimeout:       3 * time.Second,
	}
	arc.Validate()

	content, _, err := arc.Archive(ctx, req)
	if err != nil {
		return ""
	}

	name := basename(ctx) + ".htm"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, filePerm); err != nil {
		return ""
	}
	return path
}

func basename(ctx context.Context) string {
	if v, ok := ctx.Value(ctxBasenameKey).(string); ok {
		return v
	}
	return ""
}

func readOutput(rc io.ReadCloser) {
	for {
		out := make([]byte, 1024)
		_, err := rc.Read(out)
		logger.Info(string(out))
		if err != nil {
			break
		}
	}
}
