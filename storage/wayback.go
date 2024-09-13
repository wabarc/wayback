// Copyright 2024 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package storage // import "github.com/wabarc/wayback/storage"

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
)

func (s *Storage) CreateWayback(ctx context.Context, cols []wayback.Collect) error {
	if len(cols) == 0 {
		return fmt.Errorf("store: cols missing")
	}

	tx, err := s.ds.Begin()
	if err != nil {
		return fmt.Errorf("store: unable to begin transaction: %v", err)
	}

	var id int64
	query := `INSERT INTO wayback (source) VALUES ($1) RETURNING id`
	err = tx.QueryRowContext(ctx, query, cols[0].Src).Scan(&id)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return fmt.Errorf("store: unable to rollback transaction: %v", err)
		}
		return fmt.Errorf("store: unable to create wayback: %v", err)
	}

	for _, col := range cols {
		if !helper.IsURL(col.Dst) {
			continue
		}
		err = s.createArchives(ctx, tx, col, id)
		if err != nil {
			if err = tx.Rollback(); err != nil {
				return fmt.Errorf("store: unable to rollback transaction: %v", err)
			}
			return fmt.Errorf("store: create archives failed: %v", err)
		}
	}

	return tx.Commit()
}

func (s *Storage) createArchives(ctx context.Context, tx *sql.Tx, col wayback.Collect, wayback_id int64) error {
	query := `INSERT INTO archives (wayback_id, slot, dest) VALUES ($1, $2, $3)`
	_, err := tx.ExecContext(ctx, query, wayback_id, col.Arc, col.Dst)
	if err != nil {
		return fmt.Errorf("store: unable to create archive: %v", err)
	}
	return nil
}
