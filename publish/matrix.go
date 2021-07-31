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

func (m *matrixBot) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Warn("collects empty")
		return
	}

	var bnd = bundle(ctx, cols)
	var txt = render.ForPublish(&render.Matrix{Cols: cols, Data: bnd}).String()
	if m.toRoom(txt) {
		metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishMatrix, metrics.StatusFailure)
	return
}

func (m *matrixBot) toRoom(text string) bool {
	if !config.Opts.PublishToMatrixRoom() || m.client == nil {
		logger.Warn("publish to Matrix room abort.")
		return false
	}
	if text == "" {
		logger.Warn("matrix validation failed: Text can't be blank")
		return false
	}

	content := &event.MessageEventContent{
		FormattedBody: text,
		Format:        event.FormatHTML,
		Body:          text,
		MsgType:       event.MsgText,
	}
	logger.Debug("send to Matrix room, text:\n%s", text)
	if _, err := m.client.SendMessageEvent(id.RoomID(config.Opts.MatrixRoomID()), event.EventMessage, content); err != nil {
		logger.Error("send to Matrix room failure: %v", err)
		return false
	}

	return true
}
