// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"bufio"
	"embed"
	"net/url"
	"os"
	"strings"

	"github.com/wabarc/logger"
	"golang.org/x/net/publicsuffix"
)

// Copied from https://github.com/iawia002/lux/blob/f1baf46e/app/register.go#L3-L40
import (
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

const filename = "sites"

//go:embed sites
var sites embed.FS

var managedMediaSites = make(map[string]struct{})

func init() {
	parseMediaSites()
}

func baseHost(u *url.URL) (string, error) {
	dom, err := publicsuffix.EffectiveTLDPlusOne(u.Hostname())
	if err != nil {
		return "", err
	}
	return dom, nil
}

func parseMediaSites() {
	file, err := sites.Open(filename)
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
