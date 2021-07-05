// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package matrix // import "github.com/wabarc/wayback/service/matrix"

import (
	"context"
	"strings"
	"sync"

	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/template/render"
	matrix "maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Matrix struct {
	sync.RWMutex

	ctx    context.Context
	pool   pooling.Pool
	client *matrix.Client
	store  *storage.Storage
}

// New Matrix struct.
func New(ctx context.Context, store *storage.Storage, pool pooling.Pool) *Matrix {
	if config.Opts.MatrixUserID() == "" || config.Opts.MatrixPassword() == "" || config.Opts.MatrixHomeserver() == "" {
		logger.Fatal("[matrix] missing required environment variable")
	}
	if store == nil {
		logger.Fatal("[matrix] must initialize storage")
	}
	if pool == nil {
		logger.Fatal("[matrix] must initialize pooling")
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
	logger.Debug("[matrix] Serving Matrix account: %s", config.Opts.MatrixUserID())

	syncer := m.client.Syncer.(*matrix.DefaultSyncer)
	// Listen join room invite event from user
	syncer.OnEventType(event.StateMember, func(source matrix.EventSource, ev *event.Event) {
		ms := ev.Content.AsMember().Membership
		if ms == event.MembershipInvite {
			logger.Debug("[matrix] StateMember event id: %s, event type: %s, event content: %v", ev.ID, ev.Type.Type, ev.Content.Raw)
			if _, err := m.client.JoinRoomByID(ev.RoomID); err != nil {
				logger.Error("[matrix] accept invitation from sender failure, error: %v", err)
			}
		}
	})
	// Listen message event from user
	syncer.OnEventType(event.EventMessage, func(source matrix.EventSource, ev *event.Event) {
		logger.Debug("[matrix] event: %#v", ev)
		go func(ev *event.Event) {
			// Do not handle message event:
			// 1. Sent by self
			// 2. Message was deleted (when ev.Unsigned.RedactedBecause not nil)
			if ev.Sender == id.UserID(config.Opts.MatrixUserID()) || ev.Unsigned.RedactedBecause != nil {
				return
			}
			metrics.IncrementWayback(metrics.ServiceMatrix, metrics.StatusRequest)
			m.pool.Roll(func() {
				if err := m.process(ev); err != nil {
					logger.Error("[matrix] process request failure, error: %v", err)
					metrics.IncrementWayback(metrics.ServiceMatrix, metrics.StatusFailure)
				} else {
					metrics.IncrementWayback(metrics.ServiceMatrix, metrics.StatusSuccess)
				}
				// m.destroyRoom(ev.RoomID)
			})
		}(ev)
	})
	syncer.OnEventType(event.EventEncrypted, func(source matrix.EventSource, ev *event.Event) {
		logger.Error("Unsupport encryption message")
		// logger.Debug("[matrix] event: %v", ev)
		// if err := m.process(context.Background(), ev); err != nil {
		// 	logger.Error("[matrix] process request failure, error: %v", err)
		// }
		// m.destroyRoom(ev.RoomID)
	})

	go func() {
		if err := m.client.Sync(); err != nil {
			logger.Debug("[matrix] sync failed: %v", err)
		}
	}()

	<-m.ctx.Done()
	logger.Info("[matrix] stopping sync and logout all sessions")
	m.client.StopSync()
	m.client.LogoutAll()

	return errors.New("done")
}

func (m *Matrix) process(ev *event.Event) error {
	if ev.Sender == "" {
		logger.Debug("[matrix] without sender")
		return errors.New("Matrix: without sender")
	}
	logger.Debug("[matrix] event id: %s, event type: %s, event content: %v", ev.ID, ev.Type.Type, ev.Content)

	if content := ev.Content.Parsed.(*event.MessageEventContent); content.MsgType != event.MsgText {
		logger.Debug("[matrix] only support text message, current msgtype: %v", content.MsgType)
		return errors.New("Matrix: only support text message")
	}

	text := ev.Content.AsMessage().Body
	logger.Debug("[matrix] from: %s message: %s", ev.Sender, text)

	if strings.Contains(text, config.PB_SLUG) {
		return m.playback(ev)
	}

	urls := helper.MatchURLFallback(text)
	if len(urls) == 0 {
		logger.Info("[matrix] archives failure, URL no found.")
		// Redact message
		m.redact(ev, "URL no found. Original message: "+text)
		return errors.New("Matrix: URL no found")
	}

	var bundles reduxer.Bundles
	cols, err := wayback.Wayback(context.TODO(), &bundles, urls...)
	if err != nil {
		logger.Error("[matrix] archives failure, %v", err)
		return err
	}
	logger.Debug("[matrix] bundles: %#v", bundles)

	body := render.ForReply(&render.Matrix{Cols: cols}).String()
	content := &event.MessageEventContent{
		FormattedBody: body,
		Format:        event.FormatHTML,
		// Body:          body,
		// To:            id.UserID(ev.Sender),
		MsgType: event.MsgText,
	}
	content.SetReply(ev)
	if _, err := m.client.SendMessageEvent(ev.RoomID, event.EventMessage, content); err != nil {
		logger.Error("[matrix] send to Matrix room failure: %v", err)
		return err
	}
	// Redact message
	m.redact(ev, "Wayback completed. Original message: "+text)

	// Mark message as receipt
	if err := m.client.MarkRead(ev.RoomID, ev.ID); err != nil {
		logger.Error("[matrix] mark message as receipt failure: %v", err)
	}

	ctx := context.WithValue(m.ctx, publish.FlagMatrix, m.client)
	ctx = context.WithValue(ctx, publish.PubBundle, bundles)
	publish.To(ctx, cols, publish.FlagMatrix)

	return nil
}

func (m *Matrix) playback(ev *event.Event) error {
	text := ev.Content.AsMessage().Body
	urls := helper.MatchURL(text)
	// Redact message
	defer m.redact(ev, "URL no found. Original message: "+text)
	if len(urls) == 0 {
		logger.Info("[matrix] playback failure, URL no found.")
		return errors.New("Matrix: URL no found")
	}

	cols, err := wayback.Playback(m.ctx, urls...)
	if err != nil {
		logger.Error("[matrix] playback failure, %v", err)
		return err
	}

	body := render.ForReply(&render.Matrix{Cols: cols}).String()
	content := &event.MessageEventContent{
		FormattedBody: body,
		Format:        event.FormatHTML,
		MsgType:       event.MsgText,
	}
	content.SetReply(ev)
	if _, err := m.client.SendMessageEvent(ev.RoomID, event.EventMessage, content); err != nil {
		logger.Error("[matrix] send to Matrix room failure: %v", err)
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
		logger.Error("[matrix] react message failure, error: %v", err)
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
