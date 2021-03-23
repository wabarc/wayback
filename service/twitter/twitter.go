// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package twitter // import "github.com/wabarc/wayback/service/twitter"

import (
	"context"
	"sync"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/logger"
	"github.com/wabarc/wayback/publish"
)

type Twitter struct {
	sync.RWMutex

	opts   *config.Options
	client *twitter.Client

	archiving map[string]bool
}

// New returns Twitter struct.
func New(opts *config.Options) *Twitter {
	if !opts.PublishToTwitter() {
		logger.Fatal("Missing required environment variable")
	}

	config := oauth1.NewConfig(opts.TwitterConsumerKey(), opts.TwitterConsumerSecret())
	token := oauth1.NewToken(opts.TwitterAccessToken(), opts.TwitterAccessSecret())
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	return &Twitter{
		opts:   opts,
		client: client,
	}
}

// Serve loop request direct messages from the Twitter API.
// Serve always returns a nil error.
func (t *Twitter) Serve(ctx context.Context) error {
	if t.client == nil {
		return errors.New("[twitter] Initialize Twitter cilent failed.")
	}

	user, _, err := t.client.Accounts.VerifyCredentials(&twitter.AccountVerifyParams{IncludeEntities: twitter.Bool(false)})
	if err != nil {
		return errors.New("[twitter] Get account failed, err: %v", err)
	}
	logger.Info("[twitter] authorized on account %s", user.ScreenName)

	mutex := sync.RWMutex{}
	t.archiving = make(map[string]bool)
	for {
		messages, _, err := t.client.DirectMessages.EventsList(
			&twitter.DirectMessageEventsListParams{Count: 3},
		)
		logger.Debug("[twitter] messages: %v", messages)
		if err != nil {
			logger.Error("[twitter] get direct messages failed, %v", err)
		}

		for _, event := range messages.Events {
			if _, exist := t.archiving[event.ID]; exist {
				continue
			}
			go t.process(ctx, event)

			mutex.Lock()
			t.archiving[event.ID] = true
			mutex.Unlock()
		}
		time.Sleep(time.Minute)
	}
}

func (t *Twitter) process(ctx context.Context, event twitter.DirectMessageEvent) error {
	msg := event.Message
	if msg == nil || event.ID == "" {
		logger.Debug("[twitter] no direct message")
		return errors.New("Twitter: no direct message")
	}
	logger.Debug("[twitter] message event, event id: %s", event.ID)
	logger.Debug("[twitter] message event, message data: %v", msg.Data)

	text := msg.Data.Text
	logger.Debug("[twitter] message event id: %s message: %s", event.ID, text)
	defer func() {
		// Destroy Direct Message
		t.client.DirectMessages.EventsDestroy(event.ID)
		time.Sleep(time.Second)
		delete(t.archiving, event.ID)
	}()

	urls := helper.MatchURL(text)
	pub := publish.NewTwitter(t.client, t.opts)
	var realURLs []string
	for _, url := range urls {
		realURLs = append(realURLs, helper.RealURI(url))
	}
	logger.Debug("[twitter] real urls: %v", realURLs)

	if len(realURLs) == 0 {
		logger.Info("[twitter] archives failure, URL no found.")
		return errors.New("Twitter: URL no found")
	}

	col, err := t.archive(realURLs)
	if err != nil {
		logger.Error("[twitter] archives failure, ", err)
		return err
	}

	replyText := pub.Render(col)
	logger.Debug("[twitter] reply text, %s", replyText)

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
	logger.Debug("[twitter] reply event: %v", ev)
	if err != nil {
		logger.Debug("[twitter] reply error: %v", ev, err)
		return err
	}

	go func() {
		// Destroy Direct Message
		time.Sleep(time.Second)
		t.client.DirectMessages.EventsDestroy(ev.ID)
	}()

	go publish.To(ctx, t.opts, col, "twitter")

	return nil
}

func (t *Twitter) archive(urls []string) (col []*wayback.Collect, err error) {
	logger.Debug("[twitter] archives start...")

	wg := sync.WaitGroup{}
	var wbrc wayback.Broker = &wayback.Handle{URLs: urls, Opts: t.opts}
	for slot, arc := range t.opts.Slots() {
		if !arc {
			continue
		}
		wg.Add(1)
		go func(slot string) {
			defer wg.Done()
			c := &wayback.Collect{}
			logger.Debug("[twitter] archiving slot: %s", slot)
			switch slot {
			case config.SLOT_IA:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.IA()
			case config.SLOT_IS:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.IS()
			case config.SLOT_IP:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.IP()
			case config.SLOT_PH:
				c.Arc = config.SlotName(slot)
				c.Dst = wbrc.PH()
			}
			col = append(col, c)
		}(slot)
	}
	wg.Wait()

	return col, nil
}

// doc: https://developer.twitter.com/en/docs/twitter-api/v1/rate-limits
// rate limit of application/rate_limit_status: 180 requests / 15min
func (t *Twitter) isRateLimit() bool {
	rateLimits, _, err := t.client.RateLimits.Status(&twitter.RateLimitParams{Resources: []string{"statuses", "users"}})
	if err != nil {
		logger.Error("[twitter] request application/rate_limit_status failure, err: %v", err)
		return false
	}
	logger.Debug("[twitter] rate limits: %v", rateLimits)

	return true
}
