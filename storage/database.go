// Copyright 2024 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package storage // import "github.com/wabarc/wayback/storage"

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/wabarc/wayback/config"

	_ "github.com/lib/pq"
	bolt "go.etcd.io/bbolt"
)

// Open open a bolt database on current directory in given path.
// It is the caller's responsibility to close it.
func Open(opts *config.Options, path string) (*bolt.DB, error) {
	if path == "" {
		path = opts.BoltPathname()
	}
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open bolt database failed: %v", err)
	}
	return db, nil
}

// NewConnectionPool configures the database connection pool.
func NewConnectionPool(dsn string, minConnections, maxConnections int, connectionLifetime time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(minConnections)
	db.SetConnMaxLifetime(connectionLifetime)

	return db, nil
}

// Migrate executes database migrations.
// nolint: errcheck
func Migrate(db *sql.DB) error {
	var currentVersion int
	db.QueryRow(`SELECT version FROM schema_version`).Scan(&currentVersion)

	for version := currentVersion; version < schemaVersion; version++ {
		newVersion := version + 1

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}

		if err := migrations[version](tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}

		if _, err := tx.Exec(`DELETE FROM schema_version`); err != nil {
			tx.Rollback()
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}

		if _, err := tx.Exec(`INSERT INTO schema_version (version) VALUES ($1)`, newVersion); err != nil {
			tx.Rollback()
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}
	}

	return nil
}
