// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"bytes"
	"context"
	"fmt"
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
	"github.com/iawia002/lux/downloader"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/extractors/types"
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

// Bundle represents a bundle data of a webpage.
type Bundle struct {
	screenshot.Screenshots

	Assets  Assets
	Article readability.Article
}

// Assets represents the file paths stored on the local disk.
type Assets struct {
	Img, PDF, Raw, Txt, HAR, WARC, Media Asset
}

// Asset represents the files on the local disk and the remote servers.
type Asset struct {
	Local  string
	Remote Remote
}

// Remote represents the file on the remote server.
type Remote struct {
	Anonfile string
	Catbox   string
}

// Bundles represents a set of the Bundle in a map, and its key is a URL string.
type Bundles map[string]*Bundle

var (
	_, existFFmpeg       = exists("ffmpeg")
	youget, existYouGet  = exists("you-get")
	ytdl, existYoutubeDL = exists("youtube-dl")
)

// Do executes secreenshot, print PDF and export html of given URLs
// Returns a set of bundle containing screenshot data and file path
// nolint:gocyclo
func Do(ctx context.Context, urls ...*url.URL) (bundles Bundles, err error) {
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

	var wg sync.WaitGroup
	var mu sync.Mutex
	var warc = &warcraft.Warcraft{BasePath: dir, UserAgent: config.Opts.WaybackUserAgent()}
	var craft = func(in string) string {
		u, err := url.Parse(in)
		if err != nil {
			logger.Debug("create warc for %s failed", u.String())
			return ""
		}
		path, err := warc.Download(ctx, u)
		if err != nil {
			logger.Debug("create warc for %s failed: %v", u.String(), err)
			return ""
		}
		return path
	}

	type m struct {
		key *Asset
		buf []byte
	}

	for _, shot := range shots {
		wg.Add(1)
		go func(shot screenshot.Screenshots) {
			defer wg.Done()

			var assets Assets
			slugs := []m{
				{key: &assets.Img, buf: shot.Image},
				{key: &assets.PDF, buf: shot.PDF},
				{key: &assets.Raw, buf: shot.HTML},
				{key: &assets.HAR, buf: shot.HAR},
			}
			for _, slug := range slugs {
				if slug.buf == nil {
					logger.Warn("file empty, skipped")
					continue
				}
				mt := mimetype.Detect(slug.buf)
				ft := mt.String()
				fp := filepath.Join(dir, helper.FileName(shot.URL, ft))
				// Replace json with har
				if strings.HasSuffix(fp, ".json") {
					fp = strings.TrimSuffix(fp, ".json") + mt.Extension()
				}
				logger.Debug("writing file: %s", fp)
				if err := os.WriteFile(fp, slug.buf, 0o600); err != nil {
					logger.Error("write %s file failed: %v", ft, err)
					continue
				}
				if err := helper.SetField(slug.key, "Local", fp); err != nil {
					logger.Error("assign field %s to path struct failed: %v", slug.key, err)
					continue
				}
			}
			// Set path of WARC file directly to avoid read file as buffer
			if err := helper.SetField(&assets.WARC, "Local", craft(shot.URL)); err != nil {
				logger.Error("assign field WARC to path struct failed: %v", err)
			}
			if err := helper.SetField(&assets.Media, "Local", media(ctx, dir, shot.URL)); err != nil {
				logger.Error("assign field Media to path struct failed: %v", err)
			}
			u, _ := url.Parse(shot.URL)
			article, err := readability.FromReader(bytes.NewReader(shot.HTML), u)
			if err != nil {
				logger.Error("parse html failed: %v", err)
			}
			fn := strings.TrimRight(helper.FileName(shot.URL, ""), "html") + "txt"
			fp := filepath.Join(dir, fn)
			if err := os.WriteFile(fp, helper.String2Byte(article.TextContent), 0o600); err == nil && article.TextContent != "" {
				if err := helper.SetField(&assets.Txt, "Local", fp); err != nil {
					logger.Error("assign field Txt to assets struct failed: %v", err)
				}
			}
			// Upload files to third-party server
			if err := remotely(ctx, &assets); err != nil {
				logger.Error("upload files to remote server failed: %v", err)
			}
			bundle := &Bundle{shot, assets, article}
			mu.Lock()
			bundles[shot.URL] = bundle
			mu.Unlock()
		}(shot)
	}
	wg.Wait()

	return bundles, nil
}

// Capture returns screenshot.Screenshots of given URLs
func Capture(ctx context.Context, urls ...*url.URL) (shots []screenshot.Screenshots, err error) {
	opts := []screenshot.ScreenshotOption{
		screenshot.ScaleFactor(1),
		screenshot.PrintPDF(true), // print pdf
		screenshot.DumpHAR(true),  // export har
		screenshot.RawHTML(true),  // export html
		screenshot.Quality(100),   // image quality
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	shots = make([]screenshot.Screenshots, 0, len(urls))
	for _, input := range urls {
		wg.Add(1)
		go func(input *url.URL) {
			defer wg.Done()

			var shot screenshot.Screenshots
			if remote := remoteHeadless(config.Opts.ChromeRemoteAddr()); remote != nil {
				logger.Debug("reduxer using remote browser")
				addr := remote.(*net.TCPAddr)
				headless, er := screenshot.NewChromeRemoteScreenshoter(addr.String())
				if er != nil {
					logger.Error("screenshot failed: %v", er)
					return
				}
				shot, err = headless.Screenshot(ctx, input, opts...)
			} else {
				logger.Debug("reduxer using local browser")
				shot, err = screenshot.Screenshot(ctx, input, opts...)
			}
			if err != nil {
				if err == context.DeadlineExceeded {
					logger.Error("screenshot deadline: %v", err)
					return
				}
				logger.Error("screenshot error: %v", err)
				return
			}
			mu.Lock()
			shots = append(shots, shot)
			mu.Unlock()
		}(input)
	}
	wg.Wait()

	return shots, nil
}

// Asset returns paths of asset on the local disk.
func (b *Bundle) Asset() (paths []Asset) {
	if b != nil {
		logger.Debug("assets: %#v", b.Assets)
		paths = []Asset{
			b.Assets.Img,
			b.Assets.PDF,
			b.Assets.Raw,
			b.Assets.Txt,
			b.Assets.HAR,
			b.Assets.WARC,
			b.Assets.Media,
		}
	}
	return
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
	if err := os.MkdirAll(dir, 0o755); err != nil {
		logger.Error("mkdir failed: %v", err)
		return "", err
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
	fn := strings.TrimSuffix(helper.FileName(in, ""), ".html")
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

		cmd := exec.CommandContext(ctx, ytdl, args...)
		logger.Debug("youtube-dl args: %s", cmd.String())
		out, err := cmd.CombinedOutput()
		if err != nil {
			logger.Warn("start youtube-dl failed: %v", err)
		}
		logger.Debug("youtube-dl output: %s", out)

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
		cmd := exec.CommandContext(ctx, youget, args...)
		logger.Debug("youget args: %s", cmd.String())
		if err := cmd.Run(); err != nil {
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
		data, err := extractors.Extract(in, types.Options{})
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
	mtype, _ := mimetype.DetectFile(v)
	if strings.HasPrefix(mtype.String(), "video") || strings.HasPrefix(mtype.String(), "audio") {
		return v
	}

	return ""
}

func sortStreams(streams map[string]*types.Stream) []*types.Stream {
	sortedStreams := make([]*types.Stream, 0, len(streams))
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

func remotely(ctx context.Context, assets *Assets) error {
	v := []*Asset{
		&assets.Img,
		&assets.PDF,
		&assets.Raw,
		&assets.Txt,
		&assets.HAR,
		&assets.WARC,
		&assets.Media,
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
			var err error

			if !helper.Exists(asset.Local) {
				logger.Debug("local asset: %s not exists", asset.Local)
				return nil
			}
			r, e := anon.Upload(asset.Local)
			if e != nil {
				err = errors.Wrap(err, fmt.Sprintf("upload %s to anonfiles failed: %s", asset.Local, e.Error()))
			} else {
				asset.Remote.Anonfile = r.Short()
			}
			c, e := cat.Upload(asset.Local)
			if e != nil {
				err = errors.Wrap(err, fmt.Sprintf("upload %s to catbox failed: %s", asset.Local, e.Error()))
			} else {
				asset.Remote.Catbox = c
			}
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
