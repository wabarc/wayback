// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package nostr // import "github.com/wabarc/wayback/publish/nostr"

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"
	"golang.org/x/sync/errgroup"
)

// Interface guard
var _ publish.Publisher = (*Nostr)(nil)

type Nostr struct {
	ctx context.Context

	bot  *nostr.Relay
	opts *config.Options
}

// New returns a Nostr client.
func New(ctx context.Context, _ *http.Client, opts *config.Options) *Nostr {
	if !opts.PublishToNostr() {
		logger.Debug("Missing required environment variable, abort.")
		return nil
	}
	// new bot for publish is needed.
	bot := &nostr.Relay{}

	return &Nostr{ctx: ctx, bot: bot, opts: opts}
}

// Publish publish text to the Nostr of given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
func (n *Nostr) Publish(ctx context.Context, rdx reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishNostr, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishNostr, metrics.StatusFailure)
		return errors.New("publish to nostr: collects empty")
	}

	_, err := publish.Artifact(ctx, rdx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	body := render.ForPublish(&render.Nostr{Cols: cols, Data: rdx}).String()
	if err = n.publish(ctx, strings.TrimSpace(body)); err != nil {
		metrics.IncrementPublish(metrics.PublishNostr, metrics.StatusFailure)
		return errors.New("publish to nostr failed: %v", err)
	}
	metrics.IncrementPublish(metrics.PublishNostr, metrics.StatusSuccess)
	return nil
}

func (n *Nostr) publish(ctx context.Context, note string) error {
	if !n.opts.PublishToNostr() {
		return fmt.Errorf("publish to nostr abort")
	}

	if note == "" {
		return fmt.Errorf("nostr validation failed: note can't be blank")
	}
	logger.Debug("send to nostr, note:\n%s", note)

	sk := n.opts.NostrPrivateKey()
	if strings.HasPrefix(sk, "nsec") {
		if _, s, e := nip19.Decode(sk); e == nil {
			sk = s.(string)
		} else {
			return fmt.Errorf("decode private key failed: %v", e)
		}
	}
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return fmt.Errorf("failed to get public key: %v", err)
	}
	ev := nostr.Event{
		PubKey:    pk,
		Content:   note,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{},
		Kind:      nostr.KindTextNote,
	}
	if err := ev.Sign(sk); err != nil {
		return fmt.Errorf("calling sign err: %v", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	var failed int64
	for _, url := range n.opts.NostrRelayURL() {
		logger.Debug(`publish note to relay: %s`, url)
		url := url
		g.Go(func() error {
			defer func() {
				// recover from upstream panic
				if r := recover(); r != nil {
					logger.Error("publish to %s failed: %v", url, r)
				}
			}()

			relay, err := nostr.RelayConnect(ctx, url)
			if err != nil {
				return fmt.Errorf("connect to %s failed: %v", url, err)
			}

			status, err := relay.Publish(ctx, ev)
			if err != nil {
				atomic.AddInt64(&failed, 1)
				return fmt.Errorf("published to %s failed: %v", url, err)
			}
			if status != nostr.PublishStatusSucceeded {
				return fmt.Errorf("published to %s status is %s, not %s", relay, status, nostr.PublishStatusSucceeded)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		annihilated := atomic.LoadInt64(&failed) == int64(len(n.opts.NostrRelayURL()))
		if annihilated {
			return err
		}
		logger.Error("publish partially failed: %v", err)
	}

	return nil
}

// Shutdown shuts down the Nostr publish service, it always return a nil error.
func (n *Nostr) Shutdown() error {
	return nil
}
