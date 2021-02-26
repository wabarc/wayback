// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"bytes"
	"context"
	"text/template"

	mstdn "github.com/mattn/go-mastodon"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/logger"
)

type Mastodon struct {
	client *mstdn.Client
}

func NewMastodon(client *mstdn.Client, opts *config.Options) *Mastodon {
	if client == nil && opts != nil {
		client = mstdn.NewClient(&mstdn.Config{
			Server:       opts.MastodonServer(),
			ClientID:     opts.MastodonClientKey(),
			ClientSecret: opts.MastodonClientSecret(),
			AccessToken:  opts.MastodonAccessToken(),
		})
	}
	return &Mastodon{client: client}
}

func (m *Mastodon) ToMastodon(ctx context.Context, opts *config.Options, text, id string) bool {
	if !opts.PublishToMastodon() || m.client == nil {
		logger.Debug("[publish] Do not publish to Mastodon.")
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
		logger.Error("%v", err)
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
		logger.Debug("GitHub: parse template failed, %v", err)
		return ""
	}

	err = tpl.Execute(&tmplBytes, vars)
	if err != nil {
		logger.Debug("Telegram: execute template failed, %v", err)
		return ""
	}

	return tmplBytes.String()
}
