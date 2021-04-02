// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package matrix // import "github.com/wabarc/wayback/service/matrix"

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/logger"
	"github.com/wabarc/wayback/publish"
	matrix "maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Matrix struct {
	sync.RWMutex

	opts   *config.Options
	client *matrix.Client
}

// New Matrix struct.
func New(opts *config.Options) *Matrix {
	if opts.MatrixUserID() == "" || opts.MatrixPassword() == "" || opts.MatrixHomeserver() == "" {
		logger.Fatal("Missing required environment variable")
	}

	client, err := matrix.NewClient(opts.MatrixHomeserver(), "", "")
	if err != nil {
		logger.Fatal("Dial Matrix client got unpredictable error: %v", err)
	}
	_, err = client.Login(&matrix.ReqLogin{
		Type:             matrix.AuthTypePassword,
		Identifier:       matrix.UserIdentifier{Type: matrix.IdentifierTypeUser, User: opts.MatrixUserID()},
		Password:         opts.MatrixPassword(),
		StoreCredentials: true,
	})
	if err != nil {
		logger.Fatal("Login to Matrix got unpredictable error: %v", err)
	}

	return &Matrix{
		opts:   opts,
		client: client,
	}
}

// Serve loop request direct messages from the Matrix server.
// Serve returns an error.
func (m *Matrix) Serve(ctx context.Context) error {
	if m.client == nil {
		return errors.New("Must initialize Matrix client.")
	}
	logger.Debug("[matrix] Serving Matrix account: %s", m.opts.MatrixUserID())

	syncer := m.client.Syncer.(*matrix.DefaultSyncer)
	// Listen join room invite event from user
	syncer.OnEventType(event.StateMember, func(source matrix.EventSource, ev *event.Event) {
		ms := ev.Content.AsMember().Membership
		if ms == event.MembershipInvite {
			logger.Debug("[matrix] StateMember event id: %s, event type: %s, event content: %v", id.EventID(ev.ID), ev.Type.Type, ev.Content.Raw)
			if _, err := m.client.JoinRoomByID(ev.RoomID); err != nil {
				logger.Error("[matrix] accept invitation from sender failure, error: %v", err)
			}
		}
	})
	// Listen message event from user
	syncer.OnEventType(event.EventMessage, func(source matrix.EventSource, ev *event.Event) {
		logger.Debug("[matrix] event: %v", ev)
		go func(ev *event.Event) {
			if ev.Sender == id.UserID(m.opts.MatrixUserID()) {
				return
			}
			if err := m.process(ctx, ev); err != nil {
				logger.Error("[matrix] process request failure, error: %v", err)
			}
			// m.destroyRoom(ev.RoomID)
			m.redact(ev.RoomID, ev.ID)
		}(ev)
	})
	syncer.OnEventType(event.EventEncrypted, func(source matrix.EventSource, ev *event.Event) {
		logger.Error("Unsupport encryption message")
		// logger.Debug("[matrix] event: %v", ev)
		// if err := m.process(context.Background(), ev); err != nil {
		// 	logger.Error("[matrix] process request failure, error: %v", err)
		// }
		// m.destroyRoom(ev.RoomID)
		// m.redact(ev.RoomID, ev.ID)
	})

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		m.client.StopSync()
		m.client.LogoutAll()
	}()

	// Block sync
	if err := m.client.SyncWithContext(ctx); err != nil {
		return err
	}

	return errors.New("done")
}

func (m *Matrix) process(ctx context.Context, ev *event.Event) error {
	if ev.Sender == "" {
		logger.Debug("[matrix] without sender")
		return errors.New("Matrix: without sender")
	}
	logger.Debug("[matrix] event id: %s, event type: %s, event content: %v", id.EventID(ev.ID), ev.Type.Type, ev.Content)

	if content := ev.Content.Parsed.(*event.MessageEventContent); content.MsgType != event.MsgText {
		logger.Debug("[matrix] only support text message, current msgtype: %v", content.MsgType)
		return errors.New("Matrix: only support text message")
	}

	text := ev.Content.AsMessage().Body
	logger.Debug("[matrix] from: %s message: %s", ev.Sender, text)

	urls := helper.MatchURL(text)
	if len(urls) == 0 {
		logger.Info("[matrix] archives failure, URL no found.")
		return errors.New("Matrix: URL no found")
	}

	col, err := m.archive(urls)
	if err != nil {
		logger.Error("[matrix] archives failure, %v", err)
		return err
	}

	body := publish.NewMatrix(m.client, m.opts).Render(col)
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

	// Mark message as receipt
	if err := m.client.MarkRead(ev.RoomID, ev.ID); err != nil {
		logger.Error("[matrix] mark message as receipt failure: %v", err)
	}

	publish.To(ctx, m.opts, col, "matrix")

	return nil
}

func (m *Matrix) archive(urls []string) (col []*wayback.Collect, err error) {
	logger.Debug("[matrix] archives start...")

	var wg sync.WaitGroup
	var mu sync.Mutex
	var wbrc wayback.Broker = &wayback.Handle{URLs: urls, Opts: m.opts}
	for slot, arc := range m.opts.Slots() {
		if !arc {
			continue
		}
		wg.Add(1)
		go func(slot string) {
			defer wg.Done()
			c := &wayback.Collect{}
			logger.Debug("[matrix] archiving slot: %s", slot)
			switch slot {
			case config.SLOT_IA:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.IA()
			case config.SLOT_IS:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.IS()
			case config.SLOT_IP:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.IP()
			case config.SLOT_PH:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.PH()
			}
			mu.Lock()
			col = append(col, c)
			mu.Unlock()
		}(slot)
	}
	wg.Wait()

	if len(col) == 0 {
		logger.Error("[matrix] archives failure")
		return col, errors.New("archives failure")
	}
	if len(col[0].Dst) == 0 {
		logger.Error("[matrix] without results")
		return col, errors.New("without results")
	}

	return col, nil
}

func (m *Matrix) redact(roomID id.RoomID, eventID id.EventID) bool {
	if roomID == "" || m.client == nil {
		return false
	}
	extra := matrix.ReqRedact{Reason: "wayback completed."}
	if _, err := m.client.RedactEvent(roomID, eventID, extra); err != nil {
		logger.Error("[matrix] react message failure, error: %v", err)
		return false
	}
	return true
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
	if id.RoomID(m.opts.MatrixRoomID()) == roomID {
		return
	}

	m.client.LeaveRoom(roomID)
	m.client.ForgetRoom(roomID)
}
