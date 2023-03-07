// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"

	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/reduxer"
)

// Artifact returns an artifact from the reduxer that corresponds to the first
// collect in the provided slice of collects. If the artifact is not found in
// the reduxer, an error is returned.
func Artifact(ctx context.Context, rdx reduxer.Reduxer, cols []wayback.Collect) (art reduxer.Artifact, err error) {
	if len(cols) == 0 {
		return art, errors.New("no collect")
	}

	var uri = cols[0].Src
	if bundle, ok := rdx.Load(reduxer.Src(uri)); ok {
		return bundle.Artifact(), nil
	}
	return art, errors.New("reduxer data not found")
}
