// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"net/url"
	"os"
	"testing"
)

const (
	host   = `https://www.youtube.com`
	domain = `youtube.com`
)

var (
	validURL, _ = url.Parse(host)
	invalidURL  = &url.URL{Host: `invalid-tld`}
)

func TestBaseHost(t *testing.T) {
	var tests = []struct {
		url *url.URL
		exp string
	}{
		{validURL, domain},
		{invalidURL, ``},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			dom, _ := baseHost(test.url)
			if dom != test.exp {
				t.Errorf(`Unexpected extract base host, got %v instead of %v`, dom, test.exp)
			}
		})
	}
}

func TestSupportedMediaSite(t *testing.T) {
	extraDomain := "https://extra-domain.com"
	missing, _ := url.Parse("https://missing.com")
	extraURL, _ := url.Parse(extraDomain)

	var tests = []struct {
		url       *url.URL
		testname  string
		filename  string
		extra     string
		supported bool
	}{
		{validURL, `test with valid url`, filename, ``, true},
		{invalidURL, `test with invalid url`, filename, ``, false},
		{missing, `test not found`, filename, ``, false},
		{extraURL, `test extra sites`, filename, extraDomain, true},
		{invalidURL, `test extra invalid sites`, filename, extraDomain, false},
		{invalidURL, `test sites configuration file not exists`, `/path/not/exists`, extraDomain, false},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			os.Setenv("WAYBACK_MEDIA_SITES", test.extra)
			parseMediaSites(test.filename)
			supported := supportedMediaSite(test.url)
			if supported != test.supported {
				t.Errorf(`Unexpected check download media supported, got %v instead of %v`, supported, test.supported)
			}
		})
	}
}
