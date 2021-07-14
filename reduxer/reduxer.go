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

	"github.com/cixtor/readability"
	"github.com/dustin/go-humanize"
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

type Path struct {
	Img, PDF, Raw, WARC, Media string
}

type Bundle struct {
	screenshot.Screenshots

	Path    Path
	Article readability.Article
}

type Bundles map[string]Bundle

var existFFmpeg = ffmpeg()

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
			logger.Debug("[reduxer] create warc for %s failed", u.String())
			return ""
		}
		path, err := warc.Download(u)
		if err != nil {
			logger.Debug("[reduxer] create warc for %s failed: %v", u.String(), err)
			return ""
		}
		return path
	}
	var media = func(in string) string {
		if !existFFmpeg {
			logger.Warn("[reduxer] FFmpeg no found, skipped")
			return ""
		}

		data, err := extractors.Extract(in, types.Options{})
		if err != nil && len(data) == 0 {
			logger.Warn("[reduxer] data empty or error %v", err)
			return ""
		}
		dt := data[0]
		// Only download first media
		ct := string(dt.Type)
		if !strings.HasPrefix(ct, "video") {
			logger.Warn("[reduxer] resource isn't video, skipped")
			return ""
		}
		fn := strings.TrimSuffix(helper.FileName(in, ct), ".html")
		fp := filepath.Join(dir, fn) + ".mp4"
		dl := downloader.New(downloader.Options{
			OutputPath:   dir,
			OutputName:   fn,
			MultiThread:  true,
			ThreadNumber: 10,
		})
		sortedStreams := sortStreams(dt.Streams)
		if len(sortedStreams) == 0 {
			logger.Warn("[reduxer] stream not found")
			return ""
		}
		streamName := sortedStreams[0].ID
		stream, ok := dt.Streams[streamName]
		if !ok {
			logger.Warn("[reduxer] stream not found")
			return ""
		}
		logger.Debug("[reduxer] stream size: %s", humanize.Bytes(uint64(stream.Size)))
		if stream.Size > int64(config.Opts.MaxMediaSize()) {
			logger.Warn("[reduxer] video size large than %s, skipped", humanize.Bytes(config.Opts.MaxMediaSize()))
			return ""
		}
		if err := dl.Download(dt); err != nil {
			return ""
		}
		return fp
	}

	type m struct {
		key string
		buf []byte
	}

	for _, shot := range shots {
		wg.Add(1)
		go func(shot screenshot.Screenshots) {
			defer wg.Done()

			slugs := []m{
				{key: "Img", buf: shot.Image},
				{key: "PDF", buf: shot.PDF},
				{key: "Raw", buf: shot.HTML},
			}
			var path Path
			for _, slug := range slugs {
				if slug.buf == nil {
					logger.Warn("[reduxer] file empty, skipped")
					continue
				}
				ft := http.DetectContentType(slug.buf)
				fp := filepath.Join(dir, helper.FileName(shot.URL, ft))
				logger.Debug("[reduxer] writing file: %s", fp)
				if err := os.WriteFile(fp, slug.buf, 0o600); err != nil {
					logger.Error("[reduxer] write %s file failed: %v", ft, err)
					continue
				}
				if err := helper.SetField(&path, slug.key, fp); err != nil {
					logger.Error("[reduxer] assign field %s to path struct failed: %v", slug.key, err)
					continue
				}
			}
			// Set path of WARC file directly to avoid read file as buffer
			if err := helper.SetField(&path, "WARC", craft(shot.URL)); err != nil {
				logger.Error("[reduxer] assign field WARC to path struct failed: %v", err)
			}
			if err := helper.SetField(&path, "Media", media(shot.URL)); err != nil {
				logger.Error("[reduxer] assign field Media to path struct failed: %v", err)
			}
			bundle := Bundle{shot, path, readability.Article{}}
			article, err := readability.New().Parse(bytes.NewReader(shot.HTML), shot.URL)
			if err != nil {
				logger.Error("[reduxer] parse html failed: %v", err)
			}
			bundle.Article = article
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

func (b Bundle) Paths() (paths []string) {
	paths = []string{
		b.Path.Img,
		b.Path.PDF,
		b.Path.WARC,
		b.Path.Media,
	}
	return
}

func remoteHeadless(addr string) net.Addr {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		logger.Warn("[reduxer] try to connect headless browser failed: %v", err)
		return nil
	}

	if conn != nil {
		conn.Close()
		logger.Warn("[reduxer] connected: %v", conn.RemoteAddr().String())
		return conn.RemoteAddr()
	} else {
		logger.Warn("[reduxer] headless chrome don't exists")
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

func ffmpeg() bool {
	locations := []string{"ffmpeg", "ffmpeg.exe"}
	for _, path := range locations {
		found, err := exec.LookPath(path)
		if err == nil {
			return found != ""
		}
	}
	return false
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
