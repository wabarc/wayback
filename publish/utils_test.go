// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"reflect"
	"testing"

	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/reduxer"
)

func TestArtifact(t *testing.T) {
	// create a mock reduxer
	rdx := reduxer.BundleExample()
	src := "https://example.com/"
	example := reduxer.Src(src)
	bundle, ok := rdx.Load(example)
	if !ok {
		t.Fatalf("unexpected load artifat: not found")
	}

	// create some collects
	col1 := wayback.Collect{Src: src}
	col2 := wayback.Collect{Src: "https://example.org/"}

	tests := []struct {
		name string
		cols []wayback.Collect
		exp  reduxer.Artifact
		err  error
	}{
		{
			name: "valid",
			cols: []wayback.Collect{col1},
			exp:  bundle.Artifact(),
			err:  nil,
		},
		{
			name: "no collect",
			cols: []wayback.Collect{},
			exp:  reduxer.Artifact{},
			err:  errors.New("no collect"),
		},
		{
			name: "not found",
			cols: []wayback.Collect{col2},
			exp:  reduxer.Artifact{},
			err:  errors.New("reduxer data not found"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			art, err := Artifact(context.Background(), rdx, test.cols)

			if !reflect.DeepEqual(err, test.err) {
				t.Errorf("expected error %v, but got %v", test.err, err)
			}

			if art != test.exp {
				t.Errorf("expected artifact %v, but got %v", test.exp, art)
			}
		})
	}
}
