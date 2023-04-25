// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package matrix // import "github.com/wabarc/wayback/publish/matrix"

import (
	"context"
	"net/http"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	matrix "maunium.net/go/mautrix"
)

// Interface guard
var _ publish.Publisher = (*Matrix)(nil)

type Matrix struct {
	bot  *matrix.Client
	opts *config.Options
}

// New returns a Matrix client.
func New(client *http.Client, opts *config.Options) *Matrix {
	if !opts.PublishToMatrixRoom() {
		logger.Debug("Missing required environment variable, abort.")
		return nil
	}

	bot, err := matrix.NewClient(opts.MatrixHomeserver(), "", "")
	if err != nil {
		logger.Error("Dial Matrix client got unpredictable error: %v", err)
		return nil
	}
	_, err = bot.Login(&matrix.ReqLogin{
		Type:             matrix.AuthTypePassword,
		Identifier:       matrix.UserIdentifier{Type: matrix.IdentifierTypeUser, User: opts.MatrixUserID()},
		Password:         opts.MatrixPassword(),
		StoreCredentials: true,
	})
	if err != nil {
		logger.Error("Login to Matrix got unpredictable error: %v", err)
		return nil
	}
	if client != nil {
		bot.Client = client
	}

	return &Matrix{bot: bot, opts: opts}
}

// Publish publish text to the Matrix room of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (m *Matrix) Publish(ctx context.Context, rdx reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusFailure)
		return errors.New("publish to matrix: collects empty")
	}

	_, err := publish.Artifact(ctx, rdx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	var body = render.ForPublish(&render.Matrix{Cols: cols, Data: rdx}).String()
	if m.toRoom(body) {
		metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusSuccess)
		return nil
	}
	metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusFailure)
	return errors.New("publish to matrix failed")
}

func (m *Matrix) toRoom(body string) bool {
	if !m.opts.PublishToMatrixRoom() || m.bot == nil {
		logger.Warn("publish to Matrix room abort.")
		return false
	}
	if body == "" {
		logger.Warn("matrix validation failed: body can't be blank")
		return false
	}

	content := &event.MessageEventContent{
		FormattedBody: body,
		Format:        event.FormatHTML,
		Body:          body,
		MsgType:       event.MsgText,
	}
	logger.Debug("send to Matrix room, body:\n%s", body)
	if _, err := m.bot.SendMessageEvent(id.RoomID(m.opts.MatrixRoomID()), event.EventMessage, content); err != nil {
		logger.Error("send to Matrix room failure: %v", err)
		return false
	}

	return true
}

// Shutdown shuts down the Matrix publish service.
func (m *Matrix) Shutdown() error {
	if m.bot != nil {
		// Stopping sync and logout all sessions
		m.bot.StopSync()
		// nolint:errcheck
		m.bot.LogoutAll()
	}

	return nil
}
