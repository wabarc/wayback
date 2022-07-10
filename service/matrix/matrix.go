// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package matrix // import "github.com/wabarc/wayback/service/matrix"

import (
	"context"
	"strings"
	"sync"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/template/render"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	matrix "maunium.net/go/mautrix"
)

// ErrServiceClosed is returned by the Service's Serve method after a call to Shutdown.
var ErrServiceClosed = errors.New("matrix: Service closed")

// Matrix represents a Matrix service in the application
type Matrix struct {
	sync.RWMutex

	ctx    context.Context
	pool   *pooling.Pool
	client *matrix.Client
	store  *storage.Storage
}

// New Matrix struct.
func New(ctx context.Context, store *storage.Storage, pool *pooling.Pool) *Matrix {
	if config.Opts.MatrixUserID() == "" || config.Opts.MatrixPassword() == "" || config.Opts.MatrixHomeserver() == "" {
		logger.Fatal("missing required environment variable")
	}
	if store == nil {
		logger.Fatal("must initialize storage")
	}
	if pool == nil {
		logger.Fatal("must initialize pooling")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	client, err := matrix.NewClient(config.Opts.MatrixHomeserver(), "", "")
	if err != nil {
		logger.Fatal("Dial Matrix client got unpredictable error: %v", err)
	}
	_, err = client.Login(&matrix.ReqLogin{
		Type:             matrix.AuthTypePassword,
		Identifier:       matrix.UserIdentifier{Type: matrix.IdentifierTypeUser, User: config.Opts.MatrixUserID()},
		Password:         config.Opts.MatrixPassword(),
		StoreCredentials: true,
	})
	if err != nil {
		logger.Fatal("Login to Matrix got unpredictable error: %v", err)
	}

	return &Matrix{
		ctx:    ctx,
		pool:   pool,
		client: client,
		store:  store,
	}
}

// Serve loop request direct messages from the Matrix server.
// Serve returns an error.
func (m *Matrix) Serve() error {
	if m.client == nil {
		return errors.New("Must initialize Matrix client.")
	}
	logger.Warn("Serving Matrix account: %s", config.Opts.MatrixUserID())

	syncer := m.client.Syncer.(*matrix.DefaultSyncer)
	// Listen join room invite event from user
	syncer.OnEventType(event.StateMember, func(source matrix.EventSource, ev *event.Event) {
		ms := ev.Content.AsMember().Membership
		if ms == event.MembershipInvite {
			logger.Debug("StateMember event id: %s, event type: %s, event content: %v", ev.ID, ev.Type.Type, ev.Content.Raw)
			if _, err := m.client.JoinRoomByID(ev.RoomID); err != nil {
				logger.Error("accept invitation from sender failure, error: %v", err)
			}
		}
	})
	// Listen message event from user
	syncer.OnEventType(event.EventMessage, func(source matrix.EventSource, ev *event.Event) {
		logger.Debug("event: %#v", ev)
		go func(ev *event.Event) {
			// Do not handle message event:
			// 1. Sent by self
			// 2. Message was deleted (when ev.Unsigned.RedactedBecause not nil)
			if ev.Sender == id.UserID(config.Opts.MatrixUserID()) || ev.Unsigned.RedactedBecause != nil {
				return
			}
			metrics.IncrementWayback(metrics.ServiceMatrix, metrics.StatusRequest)
			bucket := pooling.Bucket{
				Request: func(ctx context.Context) error {
					if err := m.process(ctx, ev); err != nil {
						logger.Error("process request failure, error: %v", err)
						m.reply(ev, service.MsgWaybackRetrying)
						return err
					}
					metrics.IncrementWayback(metrics.ServiceMatrix, metrics.StatusSuccess)
					// m.destroyRoom(ev.RoomID)
					return nil
				},
				Fallback: func(_ context.Context) error {
					m.reply(ev, service.MsgWaybackTimeout)
					metrics.IncrementWayback(metrics.ServiceMatrix, metrics.StatusFailure)
					return nil
				},
			}
			m.pool.Put(bucket)
		}(ev)
	})
	syncer.OnEventType(event.EventEncrypted, func(source matrix.EventSource, ev *event.Event) {
		logger.Error("Unsupport encryption message")
		// logger.Debug("event: %v", ev)
		// if err := m.process(context.Background(), ev); err != nil {
		// 	logger.Error("process request failure, error: %v", err)
		// }
		// m.destroyRoom(ev.RoomID)
	})

	go func() {
		if err := m.client.Sync(); err != nil {
			logger.Warn("sync failed: %v", err)
		}
	}()

	// Block until context done
	<-m.ctx.Done()

	return ErrServiceClosed
}

// Shutdown shuts down the Matrix service, it always retuan a nil error.
func (m *Matrix) Shutdown() error {
	if m.client != nil {
		// Stopping sync and logout all sessions
		m.client.StopSync()
		m.client.LogoutAll()
	}

	return nil
}

func (m *Matrix) process(ctx context.Context, ev *event.Event) error {
	if ev.Sender == "" {
		logger.Warn("without sender")
		return errors.New("Matrix: without sender")
	}
	logger.Debug("event id: %s, event type: %s, event content: %v", ev.ID, ev.Type.Type, ev.Content)

	if content := ev.Content.Parsed.(*event.MessageEventContent); content.MsgType != event.MsgText {
		logger.Debug("only support text message, current msgtype: %v", content.MsgType)
		return errors.New("Matrix: only support text message")
	}

	text := ev.Content.AsMessage().Body
	logger.Debug("from: %s message: %s", ev.Sender, text)

	if strings.Contains(text, config.PB_SLUG) {
		return m.playback(ev)
	}

	urls := service.MatchURL(text)
	if len(urls) == 0 {
		logger.Warn("archives failure, URL no found.")
		// Redact message
		m.redact(ev, "URL no found. Original message: "+text)
		return errors.New("Matrix: URL no found")
	}

	do := func(cols []wayback.Collect, rdx reduxer.Reduxer) error {
		cols, rdx, err := wayback.Wayback(ctx, urls...)
		if err != nil {
			return errors.Wrap(err, "matrix: wayback failed")
		}
		logger.Debug("reduxer: %#v", rdx)

		body := render.ForReply(&render.Matrix{Cols: cols}).String()
		if err := m.reply(ev, body); err != nil {
			return errors.Wrap(err, "send to Matrix room failed")
		}
		// Redact message
		m.redact(ev, "wayback completed. original message: "+text)

		// Mark message as receipt
		if err := m.client.MarkRead(ev.RoomID, ev.ID); err != nil {
			logger.Error("mark message as receipt failure: %v", err)
		}

		ctx = context.WithValue(ctx, publish.FlagMatrix, m.client)
		ctx = context.WithValue(ctx, publish.PubBundle{}, rdx)
		publish.To(ctx, cols, publish.FlagMatrix.String())
		return nil
	}

	return service.Wayback(ctx, urls, do)
}

func (m *Matrix) playback(ev *event.Event) error {
	text := ev.Content.AsMessage().Body
	urls := service.MatchURL(text)
	// Redact message
	defer m.redact(ev, "URL no found. Original message: "+text)
	if len(urls) == 0 {
		logger.Warn("playback failure, URL no found.")
		return errors.New("Matrix: URL no found")
	}

	cols, err := wayback.Playback(m.ctx, urls...)
	if err != nil {
		return errors.Wrap(err, "matrix: playback failed")
	}

	body := render.ForReply(&render.Matrix{Cols: cols}).String()
	if err := m.reply(ev, body); err != nil {
		return errors.Wrap(err, "send to Matrix room failed")
	}

	return nil
}

func (m *Matrix) reply(ev *event.Event, msg string) error {
	content := &event.MessageEventContent{
		FormattedBody: msg,
		Format:        event.FormatHTML,
		MsgType:       event.MsgText,
	}
	content.SetReply(ev)
	if _, err := m.client.SendMessageEvent(ev.RoomID, event.EventMessage, content); err != nil {
		return err
	}
	return nil
}

func (m *Matrix) redact(ev *event.Event, reason string) {
	if ev.ID == "" || ev.RoomID == "" || m.client == nil {
		return
	}
	extra := matrix.ReqRedact{Reason: reason}
	if _, err := m.client.RedactEvent(ev.RoomID, ev.ID, extra); err != nil {
		logger.Error("react message failure, error: %v", err)
	}
}

func (m *Matrix) joinedRooms() []id.RoomID {
	var rooms []id.RoomID
	if m.client == nil {
		return rooms
	}
	resp, err := m.client.JoinedRooms()
	if err != nil {
		return rooms
	}

	return resp.JoinedRooms
}

func (m *Matrix) destroyRoom(roomID id.RoomID) {
	if roomID == "" || m.client == nil {
		return
	}
	if id.RoomID(config.Opts.MatrixRoomID()) == roomID {
		return
	}

	m.client.LeaveRoom(roomID)
	m.client.ForgetRoom(roomID)
}
