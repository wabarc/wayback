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

func (t *Twitter) ToTwitter(ctx context.Context, text string) bool {
	if !config.Opts.PublishToTwitter() || t.client == nil {
		logger.Debug("[publish] Do not publish to Twitter.")
		return false
	}
	if text == "" {
		logger.Info("[publish] twitter validation failed: Text can't be blank")
		return false
	}

	// TODO: character limit
	if head := title(ctx, text); head != "" {
		text = "‹ " + head + " ›\n\n" + text
	}
	tweet, resp, err := t.client.Statuses.Update(text, nil)
	if err != nil {
		logger.Error("[publish] create tweet failed: %v", err)
		return false
	}
	logger.Debug("[publish] created tweet: %v, resp: %v, err: %v", tweet, resp, err)

	return true
}

// Runder generate tweet of given wayback collects. It excluded telegra.ph
// because this link has been identified by Twitter
// nolint:stylecheck
func (m *Twitter) Render(vars []wayback.Collect) string {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}{{if not $.Arc "Telegraph"}}{{ $.Arc }}:
{{ range $src, $dst := $.Dst -}}
• {{ $dst }}
{{end}}{{end}}
{{end}}`

	tpl, err := template.New("message").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Debug("[publish] parse Twitter template failed, %v", err)
		return ""
	}

	err = tpl.Execute(&tmplBytes, vars)
	if err != nil {
		logger.Debug("[publish] execute Twitter template failed, %v", err)
		return ""
	}

	return original(vars) + strings.TrimRight(tmplBytes.String(), "\n") + "\n\n#wayback #存档"
}

func original(vars []wayback.Collect) (o string) {
	if len(vars) == 0 {
		return o
	}

	for url, _ := range vars[0].Dst {
		o += "• " + url + "\n"
	}
	if o == "" {
		return o
	}

	return "source:\n" + o + "\n====\n\n"
}
