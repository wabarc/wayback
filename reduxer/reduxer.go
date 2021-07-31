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
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-shiori/go-readability"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/extractors/types"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/screenshot"
	"github.com/wabarc/warcraft"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
)

type file struct {
	Local  string
	Remote map[string]string
}

type Assets struct {
	Img, PDF, Raw, Txt, WARC, Media file
}

type Bundle struct {
	screenshot.Screenshots

	Assets  Assets
	Article readability.Article
}

type Bundles map[string]*Bundle

var existFFmpeg = exists("ffmpeg")
var existYoutubeDL = exists("youtube-dl")

// Do executes secreenshot, print PDF and export html of given URLs
// Returns a set of bundle containing screenshot data and file path
// nolint:gocyclo
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

	var wg sync.WaitGroup
	var mu sync.Mutex
	var warc = &warcraft.Warcraft{BasePath: dir}
	var craft = func(in string) string {
		u, err := url.Parse(in)
		if err != nil {
			logger.Debug("create warc for %s failed", u.String())
			return ""
		}
		path, err := warc.Download(u)
		if err != nil {
			logger.Debug("create warc for %s failed: %v", u.String(), err)
			return ""
		}
		return path
	}

	type m struct {
		key *file
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
			}
			for _, slug := range slugs {
				if slug.buf == nil {
					logger.Warn("file empty, skipped")
					continue
				}
				ft := http.DetectContentType(slug.buf)
				fp := filepath.Join(dir, helper.FileName(shot.URL, ft))
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
			if err := os.WriteFile(fp, []byte(article.TextContent), 0o600); err == nil && article.TextContent != "" {
				if err := helper.SetField(&assets.Txt, "Local", fp); err != nil {
					logger.Error("assign field Txt to assets struct failed: %v", err)
				}
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
				logger.Error("parse url failed: %v", err)
				return
			}

			var shot screenshot.Screenshots
			if remote := remoteHeadless(config.Opts.ChromeRemoteAddr()); remote != nil {
				addr := remote.(*net.TCPAddr)
				headless, err := screenshot.NewChromeRemoteScreenshoter(addr.String())
				if err != nil {
					logger.Error("screenshot failed: %v", err)
					return
				}
				shot, err = headless.Screenshot(ctx, input, opts...)
			} else {
				shot, err = screenshot.Screenshot(ctx, input, opts...)
			}
			if err != nil {
				if err == context.DeadlineExceeded {
					logger.Error("screenshot deadline: %v", err)
					return
				}
				logger.Debug("screenshot error: %v", err)
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

func (b *Bundle) Paths() (paths []string) {
	logger.Debug("assets: %#v", b.Assets)
	paths = []string{
		b.Assets.Img.Local,
		b.Assets.PDF.Local,
		b.Assets.Raw.Local,
		b.Assets.Txt.Local,
		b.Assets.WARC.Local,
		b.Assets.Media.Local,
	}
	return
}

func remoteHeadless(addr string) net.Addr {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		logger.Warn("try to connect headless browser failed: %v", err)
		return nil
	}

	if conn != nil {
		conn.Close()
		logger.Warn("connected: %v", conn.RemoteAddr().String())
		return conn.RemoteAddr()
	} else {
		logger.Warn("headless chrome don't exists")
		return nil
	}
}

func createDir(baseDir string) (dir string, err error) {
	dir = filepath.Join(baseDir, time.Now().Format("200601"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		logger.Error("mkdir failed: %v", err)
		return "", err
	}
	return dir, nil
}

func exists(tool string) bool {
	var locations []string
	switch tool {
	case "ffmpeg":
		locations = []string{"ffmpeg", "ffmpeg.exe"}
	case "youtube-dl":
		locations = []string{"youtube-dl"}
	}

	for _, path := range locations {
		found, err := exec.LookPath(path)
		if err == nil {
			return found != ""
		}
	}

	return false
}

func media(ctx context.Context, dir, in string) string {
	logger.Debug("download media to %s, url: %s", dir, in)
	fn := strings.TrimSuffix(helper.FileName(in, ""), ".html")
	fp := filepath.Join(dir, fn)

	var viaYoutubeDL = func() string {
		if !existYoutubeDL {
			return ""
		}
		// Download media via youtube-dl
		logger.Debug("download media via youtube-dl")
		args := []string{
			"--http-chunk-size=10M", "--prefer-free-formats",
			"--no-color", "--no-cache-dir", "--no-warnings",
			"--no-progress", "--no-check-certificate",
			"--format=best[ext=mp4]/best",
			"--quiet", "--output=" + fp + ".mp4", in,
		}
		cmd := exec.CommandContext(ctx, "youtube-dl", args...)
		if err := cmd.Run(); err != nil {
			logger.Error("start youtube-dl failed: %v", err)
			return ""
		}
		paths, err := filepath.Glob(fp + "*")
		if err != nil || len(paths) == 0 {
			logger.Warn("file %s* not found", fp)
			return ""
		}
		logger.Debug("matched paths: %v", paths)
		return paths[0]
	}

	var viaAnnie = func() string {
		if !existFFmpeg {
			logger.Warn("missing FFmpeg, skipped")
			return ""
		}
		// Download media via Annie
		logger.Debug("download media via annie")
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
		v = viaAnnie()
	}
	if !helper.Exists(v) {
		logger.Warn("file %s not exists", fp)
		return ""
	}
	return v
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
