// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package entity // import "github.com/wabarc/entity"

// EntityPlayback represents a keyword for playback entity.
const EntityPlayback = "playback"

// Playback represents a Playback in the application.
type Playback struct {
	ID     uint64
	Source string
}
