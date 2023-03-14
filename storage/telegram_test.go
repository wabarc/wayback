// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package storage // import "github.com/wabarc/wayback/storage"

import (
	"path"
	"testing"

	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/entity"
)

func TestCreatePlayback(t *testing.T) {
	parser := config.NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	s, err := Open(opts, path.Join(t.TempDir(), "wayback.db"))
	if err != nil {
		t.Fatalf("Unexpected open a bolt db: %v", err)
	}
	defer s.Close()

	pb := &entity.Playback{Source: ":wayback https://example.com"}
	err = s.CreatePlayback(pb)
	if err != nil {
		t.Fatalf("Unexpected create playback, error: %v", err)
	}
}

func TestPlayback(t *testing.T) {
	parser := config.NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	s, err := Open(opts, path.Join(t.TempDir(), "wayback.db"))
	if err != nil {
		t.Fatalf("Unexpected open a bolt db: %v", err)
	}
	defer s.Close()

	dt := ":wayback https://example.com"
	pb := &entity.Playback{Source: dt}
	err = s.CreatePlayback(pb)
	if err != nil {
		t.Fatalf("Unexpected create playback, error: %v", err)
	}

	pb, err = s.Playback(pb.ID)
	if err != nil {
		t.Fatalf("Unexpected query playback, error: %v", err)
	}
	if pb.ID == 0 {
		t.Errorf("Unexpected query playback, got %d instead of grather than 0", pb.ID)
	}
	if pb.Source != dt {
		t.Errorf("Unexpected query playback, got %s instead of %s", pb.Source, dt)
	}
}

func TestRemovePlayback(t *testing.T) {
	parser := config.NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	s, err := Open(opts, path.Join(t.TempDir(), "wayback.db"))
	if err != nil {
		t.Fatalf("Unexpected open a bolt db: %v", err)
	}
	defer s.Close()

	if s.RemovePlayback(0) != nil {
		t.Error("Unexpected remove playback data")
	}
}
