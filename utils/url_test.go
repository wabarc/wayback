// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

/*
Package utils handles utils libaries for the application.
*/

package utils // import "github.com/wabarc/wayback/utils"

import (
	"strings"
	"testing"
)

func TestStrip(t *testing.T) {
	link := "https://example.com/?utm_source=wabarc&utm_medium=cpc"
	if strings.Contains(strip(link), "utm") {
		t.Fail()
	}

	link = "https://example.com/t-55534999?at_custom1=link&at_campaign=64&at_custom3=Regional+East&at_custom2=twitter&at_medium=custom7&at_custom=691F31DA-4E9E-11EB-A68F-435816F31EAE"
	if strings.Contains(strip(link), "at_custom") {
		t.Fail()
	}

	link = "https://weibointl.api.weibo.cn/share/123456.html?weibo_id=101341001431"
	if !strings.EqualFold(strip(link), "https://weibointl.api.weibo.cn/share/123456.html") {
		t.Fail()
	}
}
