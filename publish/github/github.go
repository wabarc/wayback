// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package github // import "github.com/wabarc/wayback/publish/github"

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v40/github"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"
)

// Interface guard
var _ publish.Publisher = (*GitHub)(nil)

type GitHub struct {
	ctx context.Context

	client *github.Client
	opts   *config.Options
}

// New returns a GitHub client.
func New(ctx context.Context, httpClient *http.Client, opts *config.Options) *GitHub {
	if opts.GitHubToken() == "" || opts.GitHubOwner() == "" {
		logger.Debug("GitHub personal access token is required")
		return nil
	}

	if httpClient == nil {
		// Authenticated user must grant repo:public_repo scope,
		// private repository need whole repo scope.
		auth := github.BasicAuthTransport{
			Username: opts.GitHubOwner(),
			Password: opts.GitHubToken(),
		}
		httpClient = auth.Client()
	}
	client := github.NewClient(httpClient)

	return &GitHub{ctx: ctx, client: client, opts: opts}
}

// Publish publish markdown text to the GitHub issues of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (gh *GitHub) Publish(ctx context.Context, rdx reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishGithub, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishGithub, metrics.StatusFailure)
		return errors.New("publish to github: collects empty")
	}

	_, err := publish.Artifact(ctx, rdx, cols)
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
		return nil
	}
	metrics.IncrementPublish(metrics.PublishGithub, metrics.StatusFailure)
	return errors.New("publish to github failed")
}

func (gh *GitHub) toIssues(ctx context.Context, head, body string) bool {
	if gh.client == nil {
		logger.Error("create GitHub Issues abort")
		return false
	}
	if body == "" {
		logger.Warn("github validation failed: body can't be blank")
		return false
	}

	if gh.opts.HasDebugMode() {
		user, _, _ := gh.client.Users.Get(ctx, "") // nolint:errcheck
		logger.Debug("authorized GitHub user: %v", user)
	}

	// Create an issue to GitHub
	ir := &github.IssueRequest{Title: github.String(head), Body: github.String(body)}
	issue, _, err := gh.client.Issues.Create(ctx, gh.opts.GitHubOwner(), gh.opts.GitHubRepo(), ir)
	if err != nil {
		logger.Error("create issue failed: %v", err)
		return false
	}
	logger.Debug("created issue: %v", issue)

	return true
}

// Shutdown shuts down the GitHub publish service, it always return a nil error.
func (gh *GitHub) Shutdown() error {
	return nil
}
