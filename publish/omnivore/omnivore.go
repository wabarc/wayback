// Copyright 2024 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package omnivore // import "github.com/wabarc/wayback/publish/omnivore"

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/ingress"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
)

const (
	defaultClientTimeout = 10 * time.Second
	defaultApiEndpoint   = "https://api-prod.omnivore.app/api/graphql"
)

var mutation = `
mutation SaveUrl($input: SaveUrlInput!) {
  saveUrl(input: $input) {
    ... on SaveSuccess {
      url
      clientRequestId
    }
    ... on SaveError {
      errorCodes
      message
    }
  }
}
`

// Interface guard
var _ publish.Publisher = (*Omnivore)(nil)

type Omnivore struct {
	ctx context.Context

	bot  *http.Client
	opts *config.Options
}

type errorResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type successResponse struct {
	Data struct {
		SaveUrl struct {
			Url             string `json:"url"`
			ClientRequestId string `json:"clientRequestId"`
		} `json:"saveUrl"`
	} `json:"data"`
}

// New returns a omnivore client.
func New(ctx context.Context, client *http.Client, opts *config.Options) *Omnivore {
	if opts.OmnivoreApikey() == "" {
		logger.Debug("Onmnivore integration access token is required")
		return nil
	}

	bot := ingress.Client()
	if client != nil {
		bot = client
	}
	bot.Timeout = defaultClientTimeout

	return &Omnivore{ctx: ctx, bot: bot, opts: opts}
}

// Publish save url to the Omnivore of the given cols and args.
func (o *Omnivore) Publish(_ context.Context, _ reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishOmnivore, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishOmnivore, metrics.StatusFailure)
		return errors.New("publish to omnivore: collects empty")
	}

	var payload = map[string]interface{}{
		"query": mutation,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"clientRequestId": uuid.NewString(),
				"source":          "api",
				"url":             cols[0].Src,
			},
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, defaultApiEndpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", o.opts.OmnivoreApikey())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", o.opts.WaybackUserAgent())

	resp, err := o.bot.Do(req)
	if err != nil {
		metrics.IncrementPublish(metrics.PublishOmnivore, metrics.StatusFailure)
		return err
	}

	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		metrics.IncrementPublish(metrics.PublishOmnivore, metrics.StatusFailure)
		return fmt.Errorf("omnivore: failed to parse response: %s", err)
	}

	if resp.StatusCode >= 400 {
		metrics.IncrementPublish(metrics.PublishOmnivore, metrics.StatusFailure)
		var errResponse errorResponse
		if err = json.Unmarshal(b, &errResponse); err != nil {
			return fmt.Errorf("omnivore: failed to save URL: status=%d %s", resp.StatusCode, string(b))
		}
		return fmt.Errorf("omnivore: failed to save URL: status=%d %s", resp.StatusCode, errResponse.Errors[0].Message)
	}

	var successReponse successResponse
	if err = json.Unmarshal(b, &successReponse); err != nil {
		metrics.IncrementPublish(metrics.PublishOmnivore, metrics.StatusFailure)
		return fmt.Errorf("omnivore: failed to parse response, however the request appears successful, is the url correct?: status=%d %s", resp.StatusCode, string(b))
	}

	metrics.IncrementPublish(metrics.PublishOmnivore, metrics.StatusSuccess)
	return nil
}

// Shutdown shuts down the Omnivore publish service, it always return a nil error.
func (o *Omnivore) Shutdown() error {
	return nil
}
