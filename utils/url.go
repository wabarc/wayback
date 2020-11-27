// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package utils // import "github.com/wabarc/wayback/utils"

import "regexp"

func MatchURL(text string) []string {
	re := regexp.MustCompile(`https?://(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,255}\.[a-z]{0,63}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)
	urls := []string{}
	match := re.FindAllString(text, -1)
	for _, el := range match {
		urls = append(urls, el)
	}

	return urls
}
