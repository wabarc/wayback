// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
)

func TestMatchURL(t *testing.T) {
	defer helper.CheckTest(t)

	parser := config.NewParser()
	var err error
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	var (
		u = "http://example.org"
		x = "http://example.com"
		y = "https://example.com/"
		z = "https://example.com/path"
	)

	var tests = []struct {
		text string
		leng int
	}{
		{
			text: "",
			leng: 0,
		},
		{
			text: "foo " + x,
			leng: 1,
		},
		{
			text: x + " foo " + y,
			leng: 1,
		},
		{
			text: y + " foo " + z,
			leng: 2,
		},
		{
			text: u + " foo " + x,
			leng: 2,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := len(MatchURL(test.text))
			if got != test.leng {
				t.Fatalf(`Unexpected extract URLs number from text got %d instead of %d`, got, test.leng)
			}
		})
	}
}

func TestExcludeURL(t *testing.T) {
	defer helper.CheckTest(t)

	parser := config.NewParser()
	var err error
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	var (
		u, _ = url.Parse("http://example.org")
		m, _ = url.Parse("http://t.me/s/foo")
		host = "t.me"
	)

	var tests = []struct {
		urls []*url.URL
		want []*url.URL
	}{
		{
			urls: []*url.URL{u},
			want: []*url.URL{u},
		},
		{
			urls: []*url.URL{u, m},
			want: []*url.URL{u},
		},
		{
			urls: []*url.URL{m},
			want: []*url.URL{m},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := ExcludeURL(test.urls, host)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf(`Unexpected exclude URLs number, got %v instead of %v`, got, test.want)
			}
		})
	}
}
