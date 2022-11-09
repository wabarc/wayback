// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"net/url"
	"os"
	"testing"

	"github.com/wabarc/helper"
)

func TestSupportedMediaSite(t *testing.T) {
	exist, _ := url.Parse("https://youtube.com")
	missing, _ := url.Parse("https://missing.com")

	var tests = []struct {
		url *url.URL
		exp bool
	}{
		{exist, true},
		{missing, false},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			supported := supportedMediaSite(test.url)
			if supported != test.exp {
				t.Errorf(`Unexpected check download media supported, got %v instead of %v`, supported, test.exp)
			}
		})
	}
}

func TestSupportedMediaSiteWithExtra(t *testing.T) {
	extra := "https://missing.com"
	u, _ := url.Parse(extra)

	var tests = []struct {
		url string
		exp bool
	}{
		{"", false},
		{extra, true},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			helper.Unsetenv("WAYBACK_MEDIA_SITES")
			os.Setenv("WAYBACK_MEDIA_SITES", test.url)
			parseMediaSites()
			supported := supportedMediaSite(u)
			if supported != test.exp {
				t.Errorf(`Unexpected check download media supported, got %v instead of %v`, supported, test.exp)
			}
		})
	}
}
