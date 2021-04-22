// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	mstdn "github.com/mattn/go-mastodon"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
)

type Mastodon struct {
	client *mstdn.Client
}

func NewMastodon(client *mstdn.Client) *Mastodon {
	if !config.Opts.PublishToMastodon() {
		logger.Error("Missing required environment variable")
		return new(Mastodon)
	}

	if client == nil {
		client = mstdn.NewClient(&mstdn.Config{
			Server:       config.Opts.MastodonServer(),
			ClientID:     config.Opts.MastodonClientKey(),
			ClientSecret: config.Opts.MastodonClientSecret(),
			AccessToken:  config.Opts.MastodonAccessToken(),
		})
	}

	return &Mastodon{client: client}
}

func (m *Mastodon) ToMastodon(ctx context.Context, text, id string) bool {
	if !config.Opts.PublishToMastodon() || m.client == nil {
		logger.Debug("[publish] Do not publish to Mastodon.")
		return false
	}
	if text == "" {
		logger.Info("[publish] mastodon validation failed: Text can't be blank")
		return false
	}

	// TODO: character limit
	toot := &mstdn.Toot{
		Status:     text,
		Visibility: mstdn.VisibilityPublic,
	}
	if id != "" {
		toot.InReplyToID = mstdn.ID(id)
	}
	if _, err := m.client.PostStatus(ctx, toot); err != nil {
		logger.Error("[publish] post Mastodon status failed: %v", err)
		return false
	}

	return true
}

func (m *Mastodon) Render(vars []*wayback.Collect) string {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}{{ $.Arc }}:
{{ range $src, $dst := $.Dst -}}
• {{ $dst }}
{{end}}
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

	return strings.TrimSuffix(tmplBytes.String(), "\n")
}
