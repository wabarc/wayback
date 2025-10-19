// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package storage // import "github.com/wabarc/wayback/storage"

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"time"

	bolt "go.etcd.io/bbolt"
)

var ErrDatabaseNotFound = errors.New("database not found")

// Storage handles all operations related to the database.
type Storage struct {
	db *bolt.DB
	ds *sql.DB
}

// NewStorage returns a new Storage. It is the caller's responsibility to close it.
func NewStorage(ds *sql.DB, db *bolt.DB) *Storage {
	return &Storage{db: db, ds: ds}
}

// Close the bolt database
func (s *Storage) Close() (err error) {
	if s.db != nil {
		err = errors.Join(s.db.Close(), err)
	}
	if s.ds != nil {
		err = errors.Join(s.ds.Close(), err)
	}
	if err != nil {
		return err
	}
	return nil
}

// Ping checks if the database connection works.
func (s *Storage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.ds.PingContext(ctx)
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
