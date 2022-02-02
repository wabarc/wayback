// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"net/url"
	"strings"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

// MatchURL returns a slice string contains URLs extracted from the given string.
func MatchURL(s string) (urls []*url.URL) {
	var matches []string
	if config.Opts.WaybackFallback() {
		matches = helper.MatchURLFallback(s)
	}
	matches = helper.MatchURL(s)

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
