// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/reduxer"

	discord "github.com/bwmarrin/discordgo"
	slack "github.com/slack-go/slack"
	telegram "gopkg.in/telebot.v3"
)

// MatchURL returns a slice string contains URLs extracted from the given string.
func MatchURL(opts *config.Options, s string) (urls []*url.URL) {
	matches := helper.MatchURL(s)
	if opts.WaybackFallback() {
		matches = helper.MatchURLFallback(s)
	}

	wg := sync.WaitGroup{}
	urls = make([]*url.URL, len(matches))
	for i := range matches {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			u, err := url.Parse(matches[i])
			if err != nil {
				return
			}
			urls[i] = helper.RealURI(u)
		}(i)
	}
	wg.Wait()

	return removeDuplicates(urls)
}

// ExcludeURL removes URLs based on hostname; it is only available for multiple URLs.
// The URLs given should be de-duplicated advance.
func ExcludeURL(urls []*url.URL, host string) (ex []*url.URL) {
	if len(urls) < 2 {
		return urls
	}

	for _, u := range urls {
		if u.Hostname() != host {
			ex = append(ex, u)
		}
	}
	return ex
}

func removeDuplicates(elements []*url.URL) (urls []*url.URL) {
	encountered := map[string]bool{}
	slash := "/"
	for _, u := range elements {
		key := u.User.String() + u.Host + u.Path + u.RawQuery + u.Fragment
		if u.Path == "" && !strings.HasSuffix(key, slash) {
			key += slash
		}
		if !encountered[key] {
			encountered[key] = true
			urls = append(urls, u)
		}
	}
	return
}

func filterArtifact(art reduxer.Artifact, upper int64) (paths []string) {
	assets := []reduxer.Asset{
		art.Img,
		art.PDF,
		art.Raw,
		art.Txt,
		art.HAR,
		art.HTM,
		art.WARC,
		art.Media,
	}

	var fsize int64
	for _, asset := range assets {
		if asset.Local == "" {
			continue
		}
		if !helper.Exists(asset.Local) {
			logger.Warn("invalid file %s", asset.Local)
			continue
		}
		fsize += helper.FileSize(asset.Local)
		if fsize > upper {
			logger.Warn("total file size large than %s, skipped", humanize.Bytes(uint64(upper)))
			continue
		}
		paths = append(paths, asset.Local)
	}

	return
}

// UploadToDiscord composes files that share with Discord by a given artifact.
func UploadToDiscord(opts *config.Options, rdx reduxer.Reduxer) (files []*discord.File, fn func()) {
	upper := opts.MaxAttachSize("discord")
	for _, bundle := range rdx.Bundles() {
		art := bundle.Artifact()
		for _, fp := range filterArtifact(art, upper) {
			logger.Debug("open file: %s", fp)
			rd, err := os.Open(filepath.Clean(fp))
			if err != nil {
				logger.Error("open file failed: %v", err)
				continue
			}
			files = append(files, &discord.File{Name: path.Base(fp), Reader: rd})
		}
	}

	fn = func() {
		for _, f := range files {
			f.Reader.(*os.File).Close()
		}
	}

	return
}

// UploadToSlack upload files to channel and attach as a reply by the given artifact
func UploadToSlack(client *slack.Client, opts *config.Options, art reduxer.Artifact, channel, timestamp, caption string) (err error) {
	if client == nil {
		return errors.New("client invalid")
	}

	upper := opts.MaxAttachSize("slack")
	for _, fp := range filterArtifact(art, upper) {
		rd, e := os.Open(filepath.Clean(fp))
		if e != nil {
			err = errors.Wrap(err, e.Error())
			continue
		}
		params := slack.FileUploadParameters{
			Filename:        fp,
			Reader:          rd,
			Title:           caption,
			Channels:        []string{channel},
			ThreadTimestamp: timestamp,
		}
		file, e := client.UploadFile(params)
		if e != nil {
			err = errors.Wrap(err, e.Error())
			continue
		}
		file, _, _, e = client.ShareFilePublicURL(file.ID)
		if e != nil {
			err = errors.Wrap(err, e.Error())
			continue
		}
		logger.Info("slack external file permalink: %s", file.PermalinkPublic)
	}

	return nil
}

// UploadToTelegram composes files into an album by the given artifact.
func UploadToTelegram(opts *config.Options, art reduxer.Artifact, caption string) telegram.Album {
	upper := opts.MaxAttachSize("telegram")
	var album telegram.Album
	for _, fp := range filterArtifact(art, upper) {
		logger.Debug("append document: %s", fp)
		album = append(album, &telegram.Document{
			File:     telegram.FromDisk(fp),
			Caption:  caption,
			FileName: fp,
		})
	}
	return album
}
