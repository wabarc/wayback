// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"bytes"
	"context"
	"text/template"

	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/logger"
	matrix "maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Matrix struct {
	client *matrix.Client
}

func NewMatrix(client *matrix.Client, opts *config.Options) *Matrix {
	if !opts.PublishToMatrixRoom() {
		logger.Error("Missing required environment variable, abort.")
		return new(Matrix)
	}

	if client == nil && opts != nil {
		client, err := matrix.NewClient(opts.MatrixHomeserver(), "", "")
		if err != nil {
			logger.Error("Dial Matrix client got unpredictable error: %v", err)
			return new(Matrix)
		}
		_, err = client.Login(&matrix.ReqLogin{
			Type:             matrix.AuthTypePassword,
			Identifier:       matrix.UserIdentifier{Type: matrix.IdentifierTypeUser, User: opts.MatrixUserID()},
			Password:         opts.MatrixPassword(),
			StoreCredentials: false,
		})
		if err != nil {
			logger.Error("Login to Matrix got unpredictable error: %v", err)
		}
	}

	return &Matrix{client: client}
}

func (m *Matrix) ToRoom(ctx context.Context, opts *config.Options, text string) bool {
	if !opts.PublishToMatrixRoom() || m.client == nil {
		logger.Debug("[publish] publish to Matrix room abort.")
		return false
	}
	defer func() {
		m.client.Logout()
	}()

	content := &event.MessageEventContent{
		FormattedBody: text,
		Format:        event.FormatHTML,
		Body:          text,
		MsgType:       event.MsgText,
	}
	logger.Debug("[publish] send to Matrix room, text:\n%s", text)
	if _, err := m.client.SendMessageEvent(id.RoomID(opts.MatrixRoomID()), event.EventMessage, content); err != nil {
		logger.Error("[publish] send to Matrix room failure: %v", err)
		return false
	}

	return true
}

func (m *Matrix) Render(vars []*wayback.Collect) string {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}<b>{{ $.Arc -}}</b>:<br>
{{ range $src, $dst := $.Dst -}}
• {{ $dst }}<br>
{{end}}<br>
{{end}}`

	tpl, err := template.New("message").Parse(tmpl)
	if err != nil {
		logger.Debug("[publish] parse Mastodon template failed, %v", err)
		return ""
	}

	err = tpl.Execute(&tmplBytes, vars)
	if err != nil {
		logger.Debug("[publish] execute Mastodon template failed, %v", err)
		return ""
	}

	return tmplBytes.String()
}
