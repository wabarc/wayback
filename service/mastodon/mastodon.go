// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package mastodon // import "github.com/wabarc/wayback/service/mastodon"

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-mastodon"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/template/render"
	"golang.org/x/net/html"
)

// ErrServiceClosed is returned by the Service's Serve method after a call to Shutdown.
var ErrServiceClosed = errors.New("mastodon: Service closed")

// Mastodon represents a Mastodon service in the application
type Mastodon struct {
	sync.RWMutex

	ctx    context.Context
	opts   *config.Options
	pool   *pooling.Pool
	client *mastodon.Client
	store  *storage.Storage
	pub    *publish.Publish

	archiving map[mastodon.ID]bool

	clearTick *time.Ticker
	fetchTick *time.Ticker
}

// New mastodon struct.
func New(ctx context.Context, store *storage.Storage, opts *config.Options, pool *pooling.Pool, pub *publish.Publish) *Mastodon {
	if !opts.PublishToMastodon() {
		logger.Fatal("missing required environment variable")
	}
	if store == nil {
		logger.Fatal("must initialize storage")
	}
	if opts == nil {
		logger.Fatal("must initialize options")
	}
	if pool == nil {
		logger.Fatal("must initialize pooling")
	}
	if pub == nil {
		logger.Fatal("must initialize publish")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	client := mastodon.NewClient(&mastodon.Config{
		Server:       opts.MastodonServer(),
		ClientID:     opts.MastodonClientKey(),
		ClientSecret: opts.MastodonClientSecret(),
		AccessToken:  opts.MastodonAccessToken(),
	})
	return &Mastodon{
		ctx:    ctx,
		pub:    pub,
		opts:   opts,
		pool:   pool,
		client: client,
		store:  store,
	}
}

// Serve loop request direct messages from the Mastodon instance.
// Serve always returns a nil error.
func (m *Mastodon) Serve() error {
	if m.client == nil {
		return errors.New("Must initialize Mastodon client.")
	}
	logger.Info("Serving Mastodon instance: %s", m.opts.MastodonServer())

	// rcv, err := m.client.StreamingUser(m.ctx)
	// if err != nil {
	// 	logger.Error("%v", err)
	// 	return err
	// }
	// for e := range rcv {
	// 	switch t := e.(type) {
	// 	case *mastodon.UpdateEvent:
	// 		logger.Debug("%v", t.Status)

	// 		m.status = t.Status
	// 		go m.process(m.ctx)
	// 	case *mastodon.ErrorEvent:
	// 		logger.Error("%v", e)
	// 	}
	// }

	// Clear notifications, Fetch conversations
	m.clearTick, m.fetchTick = time.NewTicker(10*time.Minute), time.NewTicker(5*time.Second)

	go func() {
		m.archiving = make(map[mastodon.ID]bool)
		for {
			select {
			case <-m.clearTick.C:
				logger.Debug("clear notifications...")
				m.client.ClearNotifications(m.ctx) // nolint:errcheck
			case <-m.fetchTick.C:
				noti, err := m.client.GetNotifications(m.ctx, nil)
				if err != nil {
					logger.Error("get notifications failed: %v", err)
				}
				logger.Debug("notifications: %v", noti)

				for _, n := range noti {
					n := n
					logger.Debug("type: %s, status: %#v", n.Type, n.Status)
					if n.Type != "mention" {
						continue
					}
					if n.Status == nil {
						continue
					}
					m.RLock()
					_, exist := m.archiving[n.Status.ID]
					m.RUnlock()
					if exist {
						continue
					}

					go func() {
						m.Lock()
						m.archiving[n.Status.ID] = true
						m.Unlock()
						metrics.IncrementWayback(metrics.ServiceMastodon, metrics.StatusRequest)
						bucket := pooling.Bucket{
							Request: func(ctx context.Context) error {
								if err := m.process(ctx, n.ID, n.Status); err != nil {
									logger.Error("process failure, notification: %#v, error: %v", n, err)
									return err
								}
								metrics.IncrementWayback(metrics.ServiceMastodon, metrics.StatusSuccess)
								return nil
							},
							Fallback: func(ctx context.Context) error {
								m.ToMastodon(ctx, service.MsgWaybackTimeout, string(n.Status.ID))
								metrics.IncrementWayback(metrics.ServiceMastodon, metrics.StatusFailure)
								return nil
							},
						}
						m.pool.Put(bucket)
						m.Lock()
						delete(m.archiving, n.ID)
						m.Unlock()
					}()
				}
			}
		}
	}()

	// Block until context done
	<-m.ctx.Done()

	return ErrServiceClosed
}

// Shutdown shuts down the Mastodon service, it always retuan a nil error.
func (m *Mastodon) Shutdown() error {
	m.clearTick.Stop()
	m.fetchTick.Stop()
	return nil
}

func (m *Mastodon) process(ctx context.Context, id mastodon.ID, status *mastodon.Status) (err error) {
	if status == nil || id == "" {
		logger.Warn("no status or conversation")
		return errors.New("Mastodon: no status or conversation")
	}
	if inReplyToID, ok := status.InReplyToID.(string); ok {
		logger.Debug("inReplyToID %s", inReplyToID)
		if status, err = m.client.GetStatus(ctx, mastodon.ID(inReplyToID)); err != nil {
			logger.Error("get status failed: %v", err)
			return err
		}
		logger.Debug("got status %#v", status)
	}

	text := textContent(status.Content)
	logger.Debug("conversation id: %s message: %s", id, text)
	defer func() {
		time.Sleep(time.Second)
		if err := m.client.DismissNotification(ctx, id); err != nil {
			logger.Warn("dismiss notification failed: %v", err)
		}
	}()

	// Process playback request if message has prefix `/playback`
	if strings.Contains(text, config.PB_SLUG) {
		return m.playback(status)
	}

	urls := service.MatchURL(m.opts, text)
	if len(urls) == 0 {
		logger.Warn("archives failure, URL no found.")
		m.ToMastodon(ctx, "URL no found", string(status.ID))
		return errors.New("Mastodon: URL no found")
	}

	do := func(cols []wayback.Collect, rdx reduxer.Reduxer) error {
		logger.Debug("reduxer: %#v", rdx)

		// Reply and publish toot as public
		m.pub.Spread(ctx, rdx, cols, publish.FlagMastodon, string(status.ID))
		return nil
	}

	return service.Wayback(ctx, m.opts, urls, do)
}

func (m *Mastodon) playback(status *mastodon.Status) error {
	text := textContent(status.Content)
	urls := service.MatchURL(m.opts, text)
	if len(urls) == 0 {
		logger.Warn("playback failure, URL no found.")
		return errors.New("Mastodon: URL no found")
	}

	cols, err := wayback.Playback(m.ctx, m.opts, urls...)
	if err != nil {
		return errors.Wrap(err, "mastodon: playback failed")
	}

	// Reply toot as public
	txt := render.ForReply(&render.Mastodon{Cols: cols}).String()
	m.ToMastodon(m.ctx, txt, string(status.ID))

	return nil
}

func (m *Mastodon) ToMastodon(ctx context.Context, text, id string) bool {
	if !m.opts.PublishToMastodon() || m.client == nil {
		logger.Warn("Do not publish to Mastodon.")
		return false
	}
	if text == "" {
		logger.Warn("mastodon validation failed: Text can't be blank")
		return false
	}

	toot := &mastodon.Toot{
		Status:     text,
		Visibility: mastodon.VisibilityPublic,
	}
	if id != "" {
		toot.InReplyToID = mastodon.ID(id)
	}
	if _, err := m.client.PostStatus(ctx, toot); err != nil {
		logger.Error("post Mastodon status failed: %v", err)
		return false
	}

	return true
}

func textContent(s string) string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	var buf bytes.Buffer

	var extractText func(node *html.Node, w *bytes.Buffer)
	extractText = func(node *html.Node, w *bytes.Buffer) {
		if node.Type == html.TextNode {
			data := strings.Trim(node.Data, "\r\n")
			if data != "" {
				w.WriteString(data)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extractText(c, w)
		}
		if node.Type == html.ElementNode {
			name := strings.ToLower(node.Data)
			if name == "br" {
				w.WriteString("\n")
			}
		}
	}
	extractText(doc, &buf)
	return buf.String()
}
