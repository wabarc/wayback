// Copyright 2024 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package datastore // import "github.com/wabarc/wayback/publish/datastore"

import (
	"context"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/storage"
)

// Interface guard
var _ publish.Publisher = (*Datastore)(nil)

type Datastore struct {
	bot  *storage.Storage
	opts *config.Options
}

// New returns a Datastore struct.
func New(store *storage.Storage, opts *config.Options) *Datastore {
	if opts.IsDefaultDatabaseURL() {
		logger.Debug("Datastore integration WAYBACK_DATABASE_URL is required")
		return nil
	}

	if store == nil {
		db, err := storage.NewConnectionPool(
			opts.DatabaseURL(),
			opts.DatabaseMinConns(),
			opts.DatabaseMaxConns(),
			opts.DatabaseConnectionLifetime(),
		)
		if err != nil {
			logger.Fatal("unable to connect to database: %v", err)
		}

		store = storage.NewStorage(db, nil)
	}

	return &Datastore{bot: store, opts: opts}
}

// Publish save url to the datastore of the given cols and args.
func (d *Datastore) Publish(ctx context.Context, _ reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishDatabase, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishDatabase, metrics.StatusFailure)
		return errors.New("publish to datastore: collects empty")
	}

	err := d.bot.CreateWayback(ctx, cols)
	if err != nil {
		metrics.IncrementPublish(metrics.PublishDatabase, metrics.StatusFailure)
		return err
	}

	metrics.IncrementPublish(metrics.PublishDatabase, metrics.StatusSuccess)
	return nil
}

// Shutdown shuts down the datastore publish service, it always return a nil error.
func (d *Datastore) Shutdown() error {
	return d.bot.Close()
}
