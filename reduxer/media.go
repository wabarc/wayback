// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"bufio"
	"context"
	"embed"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
	"golang.org/x/net/publicsuffix"
)

const filename = "sites"

//go:embed sites
var sites embed.FS

var (
	managedMediaSites = make(map[string]struct{})

	_, existFFmpeg       = exists("ffmpeg")
	youget, existYouGet  = exists("you-get")
	ytdl, existYoutubeDL = func() (string, bool) {
		if ytdlPath, ok := exists("youtube-dl"); ok {
			return ytdlPath, ok
		}
		return exists("yt-dlp")
	}()
)

func init() {
	parseMediaSites(filename)
}

func baseHost(u *url.URL) (string, error) {
	dom, err := publicsuffix.EffectiveTLDPlusOne(u.Hostname())
	if err != nil {
		return "", err
	}
	return dom, nil
}

func parseMediaSites(fn string) {
	file, err := sites.Open(fn)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		host := strings.TrimSpace(scanner.Text())
		managedMediaSites[host] = struct{}{}
	}

	// Combine extra sites
	extra := os.Getenv("WAYBACK_MEDIA_SITES")
	if len(extra) > 0 {
		for _, s := range strings.Split(extra, ",") {
			u, err := url.Parse(s)
			if err != nil {
				continue
			}
			dom, err := baseHost(u)
			if err != nil {
				continue
			}
			managedMediaSites[dom] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Warn("append managed media sites failed: %v", err)
	}
}

func supportedMediaSite(u *url.URL) bool {
	dom, err := baseHost(u)
	if err != nil {
		return false
	}
	_, ok := managedMediaSites[dom]

	return ok
}

type media struct {
	dir  string
	path string
	name string
	url  string
}

// nolint:gocyclo
func (m media) download(ctx context.Context, cfg *config.Options) string {
	logger.Debug("download media to %s, url: %s", m.dir, m.url)

	v := m.viaYoutubeDL(ctx, cfg)
	if v == "" {
		v = m.viaYouGet(ctx, cfg)
	}
	if v == "" {
		v = m.viaLux(ctx, cfg)
	}
	if !helper.Exists(v) {
		logger.Warn("file %s not exists", m.path)
		return ""
	}
	mtype, _ := mimetype.DetectFile(v) // nolint:errcheck
	if strings.HasPrefix(mtype.String(), "video") || strings.HasPrefix(mtype.String(), "audio") {
		return v
	}

	return ""
}

// Glob files by given pattern and return first file
func match(pattern string) string {
	paths, err := filepath.Glob(pattern)
	if err != nil || len(paths) == 0 {
		logger.Warn("file %s not found", pattern)
		return ""
	}
	logger.Debug("matched paths: %v", paths)
	return paths[0]
}

// Runs a command
func run(cmd *exec.Cmd, debug bool) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return err
	}
	if debug {
		readOutput(stdout)
	}

	// Wait for the process to be finished.
	// Don't care about this error in any scenario.
	// nolint:errcheck
	_ = cmd.Wait()

	return nil
}

// Download media via youtube-dl or yt-dlp
func (m media) viaYoutubeDL(ctx context.Context, cfg *config.Options) string {
	if !existYoutubeDL {
		return ""
	}
	ytdlCmd := filepath.Base(ytdl)
	logger.Debug("download media via %s", ytdlCmd)

	args := []string{
		"--http-chunk-size=10M", "--prefer-free-formats", "--restrict-filenames",
		"--rm-cache-dir", "--no-warnings",
		"--no-progress", "--no-part", "--no-mtime", "--embed-subs", "--quiet",
		"--ignore-errors", "--format=best[ext=mp4]/best", "--merge-output-format=mp4",
		"--output=" + m.path + ".%(ext)s", m.url,
	}
	if cfg.HasDebugMode() {
		args = append(args, "--verbose", "--print-traffic")
	}
	// These arguments are different between youtube-dl and yt-dlp
	if ytdlCmd == "youtube-dl" {
		args = append(args, "--no-color", "--no-check-certificate")
	} else {
		args = append(args, "--color=no_color", "--no-check-certificates")
	}

	cmd := exec.CommandContext(ctx, ytdl, args...) // nosemgrep: gitlab.gosec.G204-1
	logger.Debug("%s args: %s", ytdlCmd, cmd.String())

	if err := run(cmd, cfg.HasDebugMode()); err != nil {
		logger.Warn("start %s failed: %v", ytdlCmd, err)
	}

	return match(m.path + "*")
}

// Download media via you-get
func (m media) viaYouGet(ctx context.Context, cfg *config.Options) string {
	if !existYouGet || !existFFmpeg {
		return ""
	}
	logger.Debug("download media via you-get")
	args := []string{
		"--output-filename=" + m.path + m.url,
	}
	cmd := exec.CommandContext(ctx, youget, args...) // nosemgrep: gitlab.gosec.G204-1
	logger.Debug("youget args: %s", cmd.String())

	if err := run(cmd, cfg.HasDebugMode()); err != nil {
		logger.Warn("run you-get failed: %v", err)
	}

	return match(m.path + "*")
}

func exists(tool string) (string, bool) {
	var locations []string
	switch tool {
	case "ffmpeg":
		locations = []string{"ffmpeg", "ffmpeg.exe"}
	case "youtube-dl":
		locations = []string{"youtube-dl"}
	case "yt-dlp":
		locations = []string{"yt-dlp"}
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
