// Copyright 2024 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package storage // import "github.com/wabarc/wayback/storage"

import (
	"database/sql"
)

var schemaVersion = len(migrations)

// Order is important. Add new migrations at the end of the list.
var migrations = []func(tx *sql.Tx) error{
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TABLE schema_version (
				version text not null
			);

			CREATE TABLE wayback (
				id bigserial not null,
				source text not null,
				created_at timestamp with time zone not null default now(),
				primary key (id)
			);

			CREATE TABLE archives (
				id bigserial not null,
				wayback_id bigint not null,
				slot varchar(255) not null default '',
				dest text not null default '',
				primary key (id),
				foreign key (wayback_id) references wayback(id) on delete cascade
			);
		`
		_, err = tx.Exec(sql)
		return err
	},
}
