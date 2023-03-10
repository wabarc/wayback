// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

//go:build with_lux

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"context"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/iawia002/lux/downloader"
	"github.com/iawia002/lux/extractors"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"

	// Copied from https://github.com/iawia002/lux/blob/f1baf46e/app/register.go#L3-L40
	_ "github.com/iawia002/lux/extractors/acfun"
	_ "github.com/iawia002/lux/extractors/bcy"
	_ "github.com/iawia002/lux/extractors/bilibili"
	_ "github.com/iawia002/lux/extractors/douyin"
	_ "github.com/iawia002/lux/extractors/douyu"
	_ "github.com/iawia002/lux/extractors/eporner"
	_ "github.com/iawia002/lux/extractors/facebook"
	_ "github.com/iawia002/lux/extractors/geekbang"
	_ "github.com/iawia002/lux/extractors/haokan"
	_ "github.com/iawia002/lux/extractors/hupu"
	_ "github.com/iawia002/lux/extractors/huya"
	_ "github.com/iawia002/lux/extractors/instagram"
	_ "github.com/iawia002/lux/extractors/iqiyi"
	_ "github.com/iawia002/lux/extractors/ixigua"
	_ "github.com/iawia002/lux/extractors/kuaishou"
	_ "github.com/iawia002/lux/extractors/mgtv"
	_ "github.com/iawia002/lux/extractors/miaopai"
	_ "github.com/iawia002/lux/extractors/netease"
	_ "github.com/iawia002/lux/extractors/pixivision"
	_ "github.com/iawia002/lux/extractors/pornhub"
	_ "github.com/iawia002/lux/extractors/qq"
	_ "github.com/iawia002/lux/extractors/streamtape"
	_ "github.com/iawia002/lux/extractors/tangdou"
	_ "github.com/iawia002/lux/extractors/tiktok"
	_ "github.com/iawia002/lux/extractors/tumblr"
	_ "github.com/iawia002/lux/extractors/twitter"
	_ "github.com/iawia002/lux/extractors/udn"
	_ "github.com/iawia002/lux/extractors/universal"
	_ "github.com/iawia002/lux/extractors/vimeo"
	_ "github.com/iawia002/lux/extractors/weibo"
	_ "github.com/iawia002/lux/extractors/ximalaya"
	_ "github.com/iawia002/lux/extractors/xinpianchang"
	_ "github.com/iawia002/lux/extractors/xvideos"
	_ "github.com/iawia002/lux/extractors/yinyuetai"
	_ "github.com/iawia002/lux/extractors/youku"
	_ "github.com/iawia002/lux/extractors/youtube"
)

func (m media) viaLux(ctx context.Context, cfg *config.Options) string {
	if !existFFmpeg {
		logger.Warn("missing FFmpeg, skipped")
		return ""
	}
	// Download media via Lux
	logger.Debug("download media via lux")
	data, err := extractors.Extract(m.url, extractors.Options{})
	if err != nil || len(data) == 0 {
		logger.Warn("data empty or error %v", err)
		return ""
	}
	dt := data[0]
	dl := downloader.New(downloader.Options{
		OutputPath:   m.dir,
		OutputName:   m.name,
		MultiThread:  true,
		ThreadNumber: 10,
		ChunkSizeMB:  10,
		Silent:       !cfg.HasDebugMode(),
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
	if stream.Size > int64(cfg.MaxMediaSize()) {
		logger.Warn("media size large than %s, skipped", humanize.Bytes(cfg.MaxMediaSize()))
		return ""
	}
	if err := dl.Download(dt); err != nil {
		logger.Error("download media failed: %v", err)
		return ""
	}
	fp := m.path + "." + stream.Ext
	return fp
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
