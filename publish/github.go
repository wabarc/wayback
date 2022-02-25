// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v40/github"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/template/render"
)

type gitHub struct {
	client *github.Client
}

// NewGitHub returns a gitHub client.
func NewGitHub(httpClient *http.Client) *gitHub {
	if config.Opts.GitHubToken() == "" || config.Opts.GitHubOwner() == "" {
		logger.Error("GitHub personal access token is required")
		return new(gitHub)
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

	return &gitHub{client: client}
}

// Publish publish markdown text to the GitHub issues of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` constant.
func (gh *gitHub) Publish(ctx context.Context, cols []wayback.Collect, args ...string) {
	metrics.IncrementPublish(metrics.PublishGithub, metrics.StatusRequest)

	if len(cols) == 0 {
		logger.Warn("collects empty")
		return
	}

	rdx, _, err := extract(ctx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	var head = render.Title(cols, rdx)
	var body = render.ForPublish(&render.GitHub{Cols: cols, Data: rdx}).String()
	if head == "" {
		head = "Published at " + time.Now().Format("2006-01-02T15:04:05")
	}

	if gh.toIssues(ctx, head, body) {
		metrics.IncrementPublish(metrics.PublishGithub, metrics.StatusSuccess)
		return
	}
	metrics.IncrementPublish(metrics.PublishGithub, metrics.StatusFailure)
	return
}

func (gh *gitHub) toIssues(ctx context.Context, head, body string) bool {
	if gh.client == nil {
		logger.Error("create GitHub Issues abort")
		return false
	}
	if body == "" {
		logger.Warn("github validation failed: body can't be blank")
		return false
	}

	if config.Opts.HasDebugMode() {
		user, _, _ := gh.client.Users.Get(ctx, "")
		logger.Debug("authorized GitHub user: %v", user)
	}

	// Create an issue to GitHub
	ir := &github.IssueRequest{Title: github.String(head), Body: github.String(body)}
	issue, _, err := gh.client.Issues.Create(ctx, config.Opts.GitHubOwner(), config.Opts.GitHubRepo(), ir)
	if err != nil {
		logger.Error("create issue failed: %v", err)
		return false
	}
	logger.Debug("created issue: %v", issue)

	return true
}
