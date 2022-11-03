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
