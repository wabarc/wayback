// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package storage // import "github.com/wabarc/wayback/storage"

import (
	"encoding/binary"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"

	bolt "go.etcd.io/bbolt"
)

var ErrDatabaseNotFound = errors.New("database not found")

// Storage handles all operations related to the database.
type Storage struct {
	db *bolt.DB
}

// Open a bolt database on current directory in given path.
// It is the caller's responsibility to close it.
func Open(opts *config.Options, path string) (*Storage, error) {
	if path == "" {
		path = opts.BoltPathname()
	}
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		logger.Fatal("open bolt database failed: %v", err)
		return nil, err
	}
	return &Storage{db: db}, nil
}

// Close the bolt database
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return ErrDatabaseNotFound
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
