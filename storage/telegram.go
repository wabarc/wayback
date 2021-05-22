// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package storage // import "github.com/wabarc/wayback/storage"

import (
	"bytes"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/entity"
	bolt "go.etcd.io/bbolt"
)

func (s *Storage) createPlaybackBucket() error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.CreateBucketIfNotExists([]byte(entity.EntityPlayback))
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Playback returns playback data of the given id.
func (s *Storage) Playback(id int) (*entity.Playback, error) {
	var pb entity.Playback

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(entity.EntityPlayback))
		v := b.Get(itob(id))
		pb.Source = string(v)
		pb.ID = id
		return nil
	})

	return &pb, err
}

// CreatePlayback creates a playback callback data.
func (s *Storage) CreatePlayback(pb *entity.Playback) error {
	if err := s.createPlaybackBucket(); err != nil {
		logger.Error("[storage] create playback buckte failed: %v", err)
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket([]byte(entity.EntityPlayback))
		id, err := b.NextSequence()
		if err != nil {
			logger.Error("[storage] generate id for playback failed: %v", err)
			return err
		}
		logger.Debug("[storage] putting data to bucket, id: %d, value: %s", id, pb.Source)

		pb.ID = int(id)
		buf := bytes.NewBufferString(pb.Source).Bytes()

		return b.Put(itob(pb.ID), buf)
	})
}

// RemovePlayback removes a playback callback entry by id.
func (s *Storage) RemovePlayback(id uint64) error {
	return nil
}
