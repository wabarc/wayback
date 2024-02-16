// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package register // import "github.com/wabarc/wayback/ingress/register"

import (
	_ "github.com/wabarc/wayback/service/discord"
	_ "github.com/wabarc/wayback/service/httpd"
	_ "github.com/wabarc/wayback/service/mastodon"
	_ "github.com/wabarc/wayback/service/matrix"
	_ "github.com/wabarc/wayback/service/relaychat"
	_ "github.com/wabarc/wayback/service/slack"
	_ "github.com/wabarc/wayback/service/telegram"
	_ "github.com/wabarc/wayback/service/twitter"
	_ "github.com/wabarc/wayback/service/xmpp"
)
