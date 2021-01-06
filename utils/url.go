// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package utils // import "github.com/wabarc/wayback/utils"

import (
	"net/url"
	"regexp"
	"strings"
)

// MatchURL is extract URL from text, returns []string always.
func MatchURL(text string) []string {
	re := regexp.MustCompile(`https?://?[-a-zA-Z0-9@:%._\+~#=]{1,255}\.[a-z]{0,63}\b(?:[-a-zA-Z0-9@:%_\+.~#?&//=]*)`)
	urls := []string{}
	match := re.FindAllString(text, -1)
	for _, el := range match {
		urls = append(urls, strip(el))
	}

	return urls
}

func strip(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return ""
	}

	queries := u.Query()
	for key := range queries {
		if strings.HasPrefix(key, "utm_") || strings.HasPrefix(key, "at_custom") || strings.HasPrefix(key, "at_medium") || strings.EqualFold(key, "weibo_id") {
			queries.Del(key)
		}
	}

	u.RawQuery = queries.Encode()

	return u.String()
}
