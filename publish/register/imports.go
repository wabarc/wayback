// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package register // import "github.com/wabarc/wayback/publish/register"

import (
	_ "github.com/wabarc/wayback/publish/discord"
	_ "github.com/wabarc/wayback/publish/github"
	_ "github.com/wabarc/wayback/publish/mastodon"
	_ "github.com/wabarc/wayback/publish/matrix"
	_ "github.com/wabarc/wayback/publish/meili"
	_ "github.com/wabarc/wayback/publish/nostr"
	_ "github.com/wabarc/wayback/publish/notion"
	_ "github.com/wabarc/wayback/publish/relaychat"
	_ "github.com/wabarc/wayback/publish/slack"
	_ "github.com/wabarc/wayback/publish/telegram"
	_ "github.com/wabarc/wayback/publish/twitter"
)
