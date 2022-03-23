// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package twitter // import "github.com/wabarc/wayback/service/twitter"

import (
	"context"
	"sync"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/template/render"

	twitter "github.com/dghubble/go-twitter/twitter"
)

// ErrServiceClosed is returned by the Service's Serve method after a call to Shutdown.
var ErrServiceClosed = errors.New("twitter: Service closed")

// Twitter represents a Twitter service in the application
type Twitter struct {
	sync.RWMutex

	ctx    context.Context
	pool   *pooling.Pool
	client *twitter.Client
	store  *storage.Storage

	archiving map[string]bool

	fetchTick *time.Ticker
}

// New returns Twitter struct.
func New(ctx context.Context, store *storage.Storage, pool *pooling.Pool) *Twitter {
	if !config.Opts.PublishToTwitter() {
		logger.Fatal("missing required environment variable")
	}
	if store == nil {
		logger.Fatal("must initialize storage")
	}
	if pool == nil {
		logger.Fatal("must initialize pooling")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	oauth := oauth1.NewConfig(config.Opts.TwitterConsumerKey(), config.Opts.TwitterConsumerSecret())
	token := oauth1.NewToken(config.Opts.TwitterAccessToken(), config.Opts.TwitterAccessSecret())
	httpClient := oauth.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	return &Twitter{
		ctx:    ctx,
		pool:   pool,
		client: client,
		store:  store,
	}
}

// Serve loop request direct messages from the Twitter API.
// Serve always returns a nil error.
func (t *Twitter) Serve() error {
	if t.client == nil {
		return errors.New("Initialize Twitter cilent failed.")
	}

	user, _, err := t.client.Accounts.VerifyCredentials(&twitter.AccountVerifyParams{IncludeEntities: twitter.Bool(false)})
	if err != nil {
		return errors.New("Get account failed, err: %v", err)
	}
	logger.Info("authorized on account %s", user.ScreenName)

	t.fetchTick = time.NewTicker(time.Minute) // Fetch Direct Message event
	go func() {
		t.archiving = make(map[string]bool)
		for {
			select {
			case <-t.fetchTick.C:
				messages, resp, err := t.client.DirectMessages.EventsList(
					&twitter.DirectMessageEventsListParams{Count: 3},
				)
				logger.Debug("messages: %v", messages)
				if err != nil {
					logger.Error("get direct messages failed, %v", err)
				}
				resp.Body.Close()

				for _, event := range messages.Events {
					if _, exist := t.archiving[event.ID]; exist {
						continue
					}
					go func(event twitter.DirectMessageEvent) {
						metrics.IncrementWayback(metrics.ServiceTwitter, metrics.StatusRequest)
						t.pool.Roll(func() {
							if err := t.process(event); err != nil {
								logger.Error("process failure, message: %#v, error: %v", event.Message, err)
								metrics.IncrementWayback(metrics.ServiceTwitter, metrics.StatusFailure)
							} else {
								metrics.IncrementWayback(metrics.ServiceTwitter, metrics.StatusSuccess)
							}
						})
					}(event)

					t.Lock()
					t.archiving[event.ID] = true
					t.Unlock()
				}
			}
		}
	}()

	// Block until context done
	<-t.ctx.Done()

	return ErrServiceClosed
}

// Shutdown shuts down the Twitter service, it always retuan a nil error.
func (t *Twitter) Shutdown() error {
	t.fetchTick.Stop()
	return nil
}

func (t *Twitter) process(event twitter.DirectMessageEvent) error {
	msg := event.Message
	if msg == nil || event.ID == "" {
		logger.Warn("no direct message")
		return errors.New("Twitter: no direct message")
	}
	logger.Debug("message event, event id: %s", event.ID)
	logger.Debug("message event, message data: %v", msg.Data)

	text := msg.Data.Text
	logger.Debug("message event id: %s message: %s", event.ID, text)
	defer func() {
		// Destroy Direct Message
		resp, err := t.client.DirectMessages.EventsDestroy(event.ID)
		if err != nil {
			return
		}
		resp.Body.Close()

		time.Sleep(time.Second)
		t.Lock()
		delete(t.archiving, event.ID)
		t.Unlock()
	}()

	urls := service.MatchURL(text)
	if len(urls) == 0 {
		logger.Warn("archives failure, URL no found.")
		return errors.New("Twitter: URL no found")
	}

	cols, rdx, err := wayback.Wayback(t.ctx, urls...)
	if err != nil {
		return errors.Wrap(err, "twitter: wayback failed")
	}
	logger.Debug("reduxer: %#v", rdx)
	defer rdx.Flush()

	replyText := render.ForReply(&render.Twitter{Cols: cols}).String()
	logger.Debug("reply text, %s", replyText)

	ev, _, err := t.client.DirectMessages.EventsNew(&twitter.DirectMessageEventsNewParams{
		Event: &twitter.DirectMessageEvent{
			Type: "message_create",
			Message: &twitter.DirectMessageEventMessage{
				Target: &twitter.DirectMessageTarget{
					RecipientID: msg.SenderID,
				},
				Data: &twitter.DirectMessageData{
					Text: replyText,
				},
			},
		},
	})
	logger.Debug("reply event: %v", ev)
	if err != nil {
		logger.Error("reply error: %v", ev, err)
		return err
	}

	go func() {
		// Destroy Direct Message
		time.Sleep(time.Second)
		resp, err := t.client.DirectMessages.EventsDestroy(ev.ID)
		if err != nil {
			return
		}
		resp.Body.Close()
	}()

	ctx := context.WithValue(t.ctx, publish.FlagTwitter, t.client)
	ctx = context.WithValue(ctx, publish.PubBundle{}, rdx)
	publish.To(ctx, cols, publish.FlagTwitter.String())

	return nil
}

// doc: https://developer.twitter.com/en/docs/twitter-api/v1/rate-limits
// rate limit of application/rate_limit_status: 180 requests / 15min
// func (t *Twitter) isRateLimit() bool {
// 	rateLimits, _, err := t.client.RateLimits.Status(&twitter.RateLimitParams{Resources: []string{"statuses", "users"}})
// 	if err != nil {
// 		logger.Error("request application/rate_limit_status failure, err: %v", err)
// 		return false
// 	}
// 	logger.Debug("rate limits: %v", rateLimits)
//
// 	return true
// }
