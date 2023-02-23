// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/template/render"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	matrix "maunium.net/go/mautrix"
)

var _ Publisher = (*matrixBot)(nil)

type matrixBot struct {
	client *matrix.Client
	opts   *config.Options
}

// NewMatrix returns a matrixBot client.
func NewMatrix(client *matrix.Client, opts *config.Options) *matrixBot {
	if !opts.PublishToMatrixRoom() {
		logger.Error("Missing required environment variable, abort.")
		return new(matrixBot)
	}

	if client == nil {
		var err error
		client, err = matrix.NewClient(opts.MatrixHomeserver(), "", "")
		if err != nil {
			logger.Error("Dial Matrix client got unpredictable error: %v", err)
			return new(matrixBot)
		}
		_, err = client.Login(&matrix.ReqLogin{
			Type:             matrix.AuthTypePassword,
			Identifier:       matrix.UserIdentifier{Type: matrix.IdentifierTypeUser, User: opts.MatrixUserID()},
			Password:         opts.MatrixPassword(),
			StoreCredentials: true,
		})
		if err != nil {
			logger.Error("Login to Matrix got unpredictable error: %v", err)
		}
	}

	return &matrixBot{client: client, opts: opts}
}

// Publish publish text to the Matrix room of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (m *matrixBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusRequest)

	if len(cols) == 0 {
		return errors.New("publish to matrix: collects empty")
	}

	rdx, _, err := extract(ctx, cols)
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

func (m *matrixBot) toRoom(body string) bool {
	if !m.opts.PublishToMatrixRoom() || m.client == nil {
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
	if _, err := m.client.SendMessageEvent(id.RoomID(m.opts.MatrixRoomID()), event.EventMessage, content); err != nil {
		logger.Error("send to Matrix room failure: %v", err)
		return false
	}

	return true
}
