// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
)

type Twitter struct {
	client *twitter.Client
}

func NewTwitter(client *twitter.Client) *Twitter {
	if !config.Opts.PublishToTwitter() {
		logger.Error("Missing required environment variable")
		return new(Twitter)
	}

	if client == nil {
		oauth := oauth1.NewConfig(config.Opts.TwitterConsumerKey(), config.Opts.TwitterConsumerSecret())
		token := oauth1.NewToken(config.Opts.TwitterAccessToken(), config.Opts.TwitterAccessSecret())
		httpClient := oauth.Client(oauth1.NoContext, token)
		client = twitter.NewClient(httpClient)
	}

	return &Twitter{client: client}
}

func (t *Twitter) ToTwitter(_ context.Context, text string) bool {
	if !config.Opts.PublishToTwitter() || t.client == nil {
		logger.Debug("[publish] Do not publish to Twitter.")
		return false
	}

	// TODO: character limit
	tweet, resp, err := t.client.Statuses.Update(text, nil)
	if err != nil {
		logger.Error("[publish] create tweet failed: %v", err)
		return false
	}
	logger.Debug("[publish] created tweet: %v, resp: %v, err: %v", tweet, resp, err)

	return true
}

func (m *Twitter) Render(vars []*wayback.Collect) string {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}{{ $.Arc }}:
{{ range $src, $dst := $.Dst -}}
• {{ $dst }}
{{end}}
{{end}}`

	tpl, err := template.New("message").Parse(tmpl)
	if err != nil {
		logger.Debug("[publish] parse Twitter template failed, %v", err)
		return ""
	}

	err = tpl.Execute(&tmplBytes, vars)
	if err != nil {
		logger.Debug("[publish] execute Twitter template failed, %v", err)
		return ""
	}

	return strings.TrimSuffix(tmplBytes.String(), "\n")
}
