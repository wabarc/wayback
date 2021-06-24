// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"bytes"
	"context"
	"net/http"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
)

type GitHub struct {
	client *github.Client
}

func NewGitHub(httpClient *http.Client) *GitHub {
	if config.Opts.GitHubToken() == "" || config.Opts.GitHubOwner() == "" {
		logger.Fatal("[publish] GitHub personal access token is required")
	}

	if httpClient == nil {
		// Authenticated user must grant repo:public_repo scope,
		// private repository need whole repo scope.
		auth := github.BasicAuthTransport{
			Username: config.Opts.GitHubOwner(),
			Password: config.Opts.GitHubToken(),
		}
		httpClient = auth.Client()
	}
	client := github.NewClient(httpClient)

	return &GitHub{
		client: client,
	}
}

func (gh *GitHub) ToIssues(ctx context.Context, text string) bool {
	if gh.client == nil {
		logger.Error("[publish] create GitHub Issues abort")
		return false
	}
	if text == "" {
		logger.Info("[publish] github validation failed: Text can't be blank")
		return false
	}

	if config.Opts.HasDebugMode() {
		user, _, _ := gh.client.Users.Get(ctx, "")
		logger.Debug("[publish] authorized GitHub user: %v", user)
	}

	title := func(s string) string {
		regex := regexp.MustCompile(`\(https:\/\/telegra\.ph\/(.*?)-\d{2}-\d{2}`)
		match := regex.FindAllStringSubmatch(s, -1)
		words := ""
		for _, m := range match {
			if len(m) == 2 {
				words += m[1] + "\t"
			}
		}
		title := []rune(words)
		limit := len(title)
		switch {
		case limit > 256:
			title = title[:256]
		case limit == 0:
			title = []rune("Published at " + time.Now().Format("2006-01-02T15:04:05"))
		case limit > 0:
		}
		return strings.TrimSpace(string(title))
	}

	// Create an issue to GitHub
	ir := &github.IssueRequest{Title: github.String(title(text)), Body: github.String(text)}
	issue, _, err := gh.client.Issues.Create(ctx, config.Opts.GitHubOwner(), config.Opts.GitHubRepo(), ir)
	if err != nil {
		logger.Debug("[publish] create issue failed: %v", err)
		return false
	}
	logger.Debug("[publish] created issue: %v", issue)

	return true
}

func (gh *GitHub) Render(vars []wayback.Collect) string {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}**[{{ $.Arc }}]({{ $.Ext }})**:
{{ range $src, $dst := $.Dst -}}
> origin: [{{ $src | unescape | revert }}]({{ $src | revert }})
> archived: {{ if $dst | isURL }}[{{ $dst | unescape }}]({{ $dst }}){{ else }}{{ $dst }}{{ end }}
{{end}}
{{end}}`

	tpl, err := template.New("message").Funcs(funcMap()).Parse(tmpl)
	if err != nil {
		logger.Debug("[publish] parse template failed, %v", err)
		return ""
	}

	err = tpl.Execute(&tmplBytes, vars)
	if err != nil {
		logger.Debug("[publish] execute template failed, %v", err)
		return ""
	}

	return strings.TrimSuffix(tmplBytes.String(), "\n")
}
