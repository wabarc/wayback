// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"

	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/template/render"
	matrix "maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type matrixBot struct {
	client *matrix.Client
}

// NewMatrix returns a matrixBot client.
func NewMatrix(client *matrix.Client) *matrixBot {
	if !config.Opts.PublishToMatrixRoom() {
		logger.Error("Missing required environment variable, abort.")
		return new(matrixBot)
	}

	if client == nil {
		var err error
		client, err = matrix.NewClient(config.Opts.MatrixHomeserver(), "", "")
		if err != nil {
			logger.Error("Dial Matrix client got unpredictable error: %v", err)
			return new(matrixBot)
		}
		_, err = client.Login(&matrix.ReqLogin{
			Type:             matrix.AuthTypePassword,
			Identifier:       matrix.UserIdentifier{Type: matrix.IdentifierTypeUser, User: config.Opts.MatrixUserID()},
			Password:         config.Opts.MatrixPassword(),
			StoreCredentials: true,
		})
		if err != nil {
			logger.Error("Login to Matrix got unpredictable error: %v", err)
		}
	}

	return &matrixBot{client: client}
}

// Publish publish text to the Matrix room of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (m *matrixBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Warn("collects empty")
		return
	}

	rdx, _, err := extract(ctx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	var body = render.ForPublish(&render.Matrix{Cols: cols, Data: rdx}).String()
	if m.toRoom(body) {
		metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusFailure)
	return
}

func (m *matrixBot) toRoom(body string) bool {
	if !config.Opts.PublishToMatrixRoom() || m.client == nil {
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
	if _, err := m.client.SendMessageEvent(id.RoomID(config.Opts.MatrixRoomID()), event.EventMessage, content); err != nil {
		logger.Error("send to Matrix room failure: %v", err)
		return false
	}

	return true
}
