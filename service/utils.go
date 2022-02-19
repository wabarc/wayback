// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"net/url"
	"os"
	"path"
	"strings"

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
func MatchURL(s string) (urls []*url.URL) {
	matches := helper.MatchURL(s)
	if config.Opts.WaybackFallback() {
		matches = helper.MatchURLFallback(s)
	}

	for i := range matches {
		u, _ := url.Parse(matches[i])
		urls = append(urls, u)
	}

	return removeDuplicates(urls)
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

// UploadToDiscord composes files that share with Discord by a given bundle.
func UploadToDiscord(bundle *reduxer.Bundle) (files []*discord.File) {
	if bundle != nil {
		var fsize int64
		upper := config.Opts.MaxAttachSize("discord")
		for _, asset := range bundle.Asset() {
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
			logger.Debug("open file: %s", asset.Local)
			rd, err := os.Open(asset.Local)
			if err != nil {
				logger.Error("open file failed: %v", err)
				continue
			}
			files = append(files, &discord.File{Name: path.Base(asset.Local), Reader: rd})
		}
	}
	return
}

// UploadToSlack upload files to channel and attach as a reply by the given bundle
func UploadToSlack(client *slack.Client, bundle *reduxer.Bundle, channel, timestamp string) (err error) {
	if client == nil {
		return errors.New("client invalid")
	}

	var fsize int64
	for _, asset := range bundle.Asset() {
		if asset.Local == "" {
			continue
		}
		if !helper.Exists(asset.Local) {
			err = errors.Wrap(err, "invalid file "+asset.Local)
			continue
		}
		fsize += helper.FileSize(asset.Local)
		if fsize > config.Opts.MaxAttachSize("slack") {
			err = errors.Wrap(err, "total file size large than 5GB, skipped")
			continue
		}
		reader, e := os.Open(asset.Local)
		if e != nil {
			err = errors.Wrap(err, e.Error())
			continue
		}
		params := slack.FileUploadParameters{
			Filename:        asset.Local,
			Reader:          reader,
			Title:           bundle.Title,
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

// UploadToTelegram composes files into an album by the given bundle.
func UploadToTelegram(bundle *reduxer.Bundle) telegram.Album {
	// Attach image and pdf files
	var album telegram.Album
	var fsize int64
	for _, asset := range bundle.Asset() {
		if asset.Local == "" {
			continue
		}
		if !helper.Exists(asset.Local) {
			logger.Warn("invalid file %s", asset.Local)
			continue
		}
		fsize += helper.FileSize(asset.Local)
		if fsize > config.Opts.MaxAttachSize("telegram") {
			logger.Warn("total file size large than 50MB, skipped")
			continue
		}
		logger.Debug("append document: %s", asset.Local)
		album = append(album, &telegram.Document{
			File:     telegram.FromDisk(asset.Local),
			Caption:  bundle.Title,
			FileName: asset.Local,
		})
	}
	return album
}
