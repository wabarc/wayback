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
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-shiori/go-readability"
	"github.com/go-shiori/obelisk"
	"github.com/iawia002/lux/downloader"
	"github.com/iawia002/lux/extractors"
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

	_, existFFmpeg       = exists("ffmpeg")
	youget, existYouGet  = exists("you-get")
	ytdl, existYoutubeDL = exists("youtube-dl")
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
func Do(ctx context.Context, urls ...*url.URL) (Reduxer, error) {
	// Returns an initialized Reduxer for safe.
	var bs = NewReduxer()
	var err error

	if !config.Opts.EnabledReduxer() {
		return bs, errors.New("Specify directory to environment `WAYBACK_STORAGE_DIR` to enable reduxer")
	}

	dir, err := createDir(config.Opts.StorageDir())
	if err != nil {
		return bs, errors.Wrap(err, "create storage directory failed")
	}

	var warc = &warcraft.Warcraft{BasePath: dir, UserAgent: config.Opts.WaybackUserAgent()}
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

			shot, er := capture(ctx, uri, dir)
			if er != nil {
				return errors.Wrap(er, "capture failed")
			}

			artifact := &Artifact{
				Img:  Asset{Local: fmt.Sprint(shot.Image)},
				Raw:  Asset{Local: fmt.Sprint(shot.HTML)},
				PDF:  Asset{Local: fmt.Sprint(shot.PDF)},
				HAR:  Asset{Local: fmt.Sprint(shot.HAR)},
				WARC: Asset{Local: craft(uri)},
			}

			if supportedMediaSite(uri) {
				artifact.Media.Local = media(ctx, dir, shot.URL)
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
			fp := filepath.Join(dir, txtName)
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
func capture(ctx context.Context, uri *url.URL, dir string) (shot *screenshot.Screenshots[screenshot.Path], err error) {
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

	if remote := remoteHeadless(config.Opts.ChromeRemoteAddr()); remote != nil {
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

func exists(tool string) (string, bool) {
	var locations []string
	switch tool {
	case "ffmpeg":
		locations = []string{"ffmpeg", "ffmpeg.exe"}
	case "youtube-dl":
		locations = []string{"youtube-dl"}
	case "you-get":
		locations = []string{"you-get"}
	}

	for _, path := range locations {
		found, err := exec.LookPath(path)
		if err == nil {
			return found, found != ""
		}
	}

	return "", false
}

// nolint:gocyclo
func media(ctx context.Context, dir, in string) string {
	logger.Debug("download media to %s, url: %s", dir, in)
	fn := basename(ctx)
	fp := filepath.Join(dir, fn)

	// Glob files by given pattern and return first file
	var match = func(pattern string) string {
		paths, err := filepath.Glob(pattern)
		if err != nil || len(paths) == 0 {
			logger.Warn("file %s* not found", fp)
			return ""
		}
		logger.Debug("matched paths: %v", paths)
		return paths[0]
	}

	// Runs a command
	var run = func(cmd *exec.Cmd) error {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		cmd.Stderr = cmd.Stdout
		if err := cmd.Start(); err != nil {
			return err
		}
		if config.Opts.HasDebugMode() {
			readOutput(stdout)
		}

		// Wait for the process to be finished.
		// Don't care about this error in any scenario.
		// nolint:errcheck
		_ = cmd.Wait()

		return nil
	}

	// Download media via youtube-dl
	var viaYoutubeDL = func() string {
		if !existYoutubeDL {
			return ""
		}
		logger.Debug("download media via youtube-dl")

		args := []string{
			"--http-chunk-size=10M", "--prefer-free-formats", "--restrict-filenames",
			"--no-color", "--rm-cache-dir", "--no-warnings", "--no-check-certificate",
			"--no-progress", "--no-part", "--no-mtime", "--embed-subs", "--quiet",
			"--ignore-errors", "--format=best[ext=mp4]/best", "--merge-output-format=mp4",
			"--output=" + fp + ".%(ext)s", in,
		}
		if config.Opts.HasDebugMode() {
			args = append(args, "--verbose", "--print-traffic")
		}

		cmd := exec.CommandContext(ctx, ytdl, args...) // nosemgrep: gitlab.gosec.G204-1
		logger.Debug("youtube-dl args: %s", cmd.String())

		if err := run(cmd); err != nil {
			logger.Warn("start youtube-dl failed: %v", err)
		}

		return match(fp + "*")
	}

	// Download media via you-get
	var viaYouGet = func() string {
		if !existYouGet || !existFFmpeg {
			return ""
		}
		logger.Debug("download media via you-get")
		args := []string{
			"--output-filename=" + fp, in,
		}
		cmd := exec.CommandContext(ctx, youget, args...) // nosemgrep: gitlab.gosec.G204-1
		logger.Debug("youget args: %s", cmd.String())

		if err := run(cmd); err != nil {
			logger.Warn("run you-get failed: %v", err)
		}

		return match(fp + "*")
	}

	var viaLux = func() string {
		if !existFFmpeg {
			logger.Warn("missing FFmpeg, skipped")
			return ""
		}
		// Download media via Lux
		logger.Debug("download media via lux")
		data, err := extractors.Extract(in, extractors.Options{})
		if err != nil || len(data) == 0 {
			logger.Warn("data empty or error %v", err)
			return ""
		}
		dt := data[0]
		dl := downloader.New(downloader.Options{
			OutputPath:   dir,
			OutputName:   fn,
			MultiThread:  true,
			ThreadNumber: 10,
			ChunkSizeMB:  10,
			Silent:       !config.Opts.HasDebugMode(),
		})
		sortedStreams := sortStreams(dt.Streams)
		if len(sortedStreams) == 0 {
			logger.Warn("stream not found")
			return ""
		}
		streamName := sortedStreams[0].ID
		stream, ok := dt.Streams[streamName]
		if !ok {
			logger.Warn("stream not found")
			return ""
		}
		logger.Debug("stream size: %s", humanize.Bytes(uint64(stream.Size)))
		if stream.Size > int64(config.Opts.MaxMediaSize()) {
			logger.Warn("media size large than %s, skipped", humanize.Bytes(config.Opts.MaxMediaSize()))
			return ""
		}
		if err := dl.Download(dt); err != nil {
			logger.Error("download media failed: %v", err)
			return ""
		}
		fp += "." + stream.Ext
		return fp
	}

	v := viaYoutubeDL()
	if v == "" {
		v = viaYouGet()
	}
	if v == "" {
		v = viaLux()
	}
	if !helper.Exists(v) {
		logger.Warn("file %s not exists", fp)
		return ""
	}
	mtype, _ := mimetype.DetectFile(v) // nolint:errcheck
	if strings.HasPrefix(mtype.String(), "video") || strings.HasPrefix(mtype.String(), "audio") {
		return v
	}

	return ""
}

func sortStreams(streams map[string]*extractors.Stream) []*extractors.Stream {
	sortedStreams := make([]*extractors.Stream, 0, len(streams))
	for _, data := range streams {
		sortedStreams = append(sortedStreams, data)
	}
	if len(sortedStreams) > 1 {
		sort.Slice(
			sortedStreams, func(i, j int) bool { return sortedStreams[i].Size > sortedStreams[j].Size },
		)
	}
	return sortedStreams
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
