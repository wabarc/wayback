// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package slack // import "github.com/wabarc/wayback/service/slack"

import (
	"context"
	"net/url"

	"github.com/gookit/color"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
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
)

var callbackKey = "playback"

// Interface guard
var _ service.Servicer = (*Slack)(nil)

// ErrServiceClosed is returned by the Service's Serve method after a call to Shutdown.
var ErrServiceClosed = errors.New("slack: Service closed")

// Slack handles a slack service.
//
// Steps to create a bot:
//
// 1. Create an App
//
// 2. Generate an App-Level Token, scopes: `connections:write`
//
// 3. Enable Socket Mode
//
// 4. Enable Events
// Subscribe to bot events: `app_mention` and `message.im`,
// Subscribe to events on behalf of users: `message.im`
//
// 5. Setting OAuth & Permissions
// User Token Scopes: `chat:write`, `files:write`
//
// 6. Install to Workspace, got `Bot User OAuth Token`
//
// 7. App Home, check `Allow users to send Slash commands and messages from the messages tab`
//
// TODO: rate limit
type Slack struct {
	ctx context.Context

	bot    *slack.Client
	client *socketmode.Client
	store  *storage.Storage
	opts   *config.Options
	pool   *pooling.Pool
	pub    *publish.Publish
}

type event struct {
	User, Text, Channel, TimeStamp, ThreadTimeStamp string
}

// New Slack struct.
func New(ctx context.Context, opts service.Options) (*Slack, error) {
	if !opts.Config.SlackEnabled() {
		return nil, errors.New("missing required environment variable, skipped")
	}
	bot := slack.New(
		opts.Config.SlackBotToken(),
		// slack.Config.OptionDebug(opts.Config.HasDebugMode()),
		slack.OptionAppLevelToken(opts.Config.SlackAppToken()),
	)
	if bot == nil {
		return nil, errors.New("create slack bot instance failed")
	}

	client := socketmode.New(
		bot,
		// socketmode.OptionDebug(opts.Config.HasDebugMode()),
		// socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	if ctx == nil {
		ctx = context.Background()
	}

	return &Slack{
		ctx:    ctx,
		bot:    bot,
		client: client,
		store:  opts.Storage,
		opts:   opts.Config,
		pool:   opts.Pool,
		pub:    opts.Publish,
	}, nil
}

// Serve loop request message from the Slack api server.
// Serve always returns an error.
func (s *Slack) Serve() (err error) {
	if s.bot == nil {
		return errors.New("Initialize slack failed, error: %v", err)
	}
	user, err := s.bot.AuthTest()
	if err != nil {
		return err
	}
	logger.Info("authorized on account %s", color.Blue.Sprint(user.User))

	go func() {
		for evt := range s.client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				logger.Info("connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				logger.Warn("connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				logger.Info("connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				s.handleRequest(evt)
			case socketmode.EventTypeInteractive:
				s.handleButton(evt)
			case socketmode.EventTypeSlashCommand:
				s.handleCommand(evt)
			case socketmode.EventTypeHello, socketmode.EventTypeDisconnect, socketmode.EventTypeIncomingError:
			default:
				logger.Warn("unexpected event type received: %s", evt.Type)
			}
		}
	}()

	logger.Info("starting slack service...")
	// Block until context done
	if err = s.client.RunContext(s.ctx); err != nil && err != context.Canceled {
		return err
	}

	return ErrServiceClosed
}

// Shutdown shuts down the Slack service, it always retuan a nil error.
func (s *Slack) Shutdown() error {
	return nil
}

func (s *Slack) handleRequest(evt socketmode.Event) {
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		logger.Warn("unsupported event: %+v", evt)
		return
	}
	logger.Debug("event received: %+v", eventsAPIEvent)

	s.client.Ack(*evt.Request)

	switch eventsAPIEvent.Type {
	case slackevents.CallbackEvent:
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			logger.Debug("channel mention message event: %+v", ev)
			// nolint:errcheck
			go s.process(&event{ev.User, ev.Text, ev.Channel, ev.TimeStamp, ev.ThreadTimeStamp})
		case *slackevents.MessageEvent:
			logger.Debug("direct message event: %+v", ev)
			// Message event https://api.slack.com/events/message
			// Exclude message from bot, https://api.slack.com/events/message/bot_message
			// Exclude message changed event
			if ev.BotID != "" || ev.SubType != "" {
				logger.Debug("skipped event from bot")
				return
			}
			// nolint:errcheck
			go s.process(&event{ev.User, ev.Text, ev.Channel, ev.TimeStamp, ev.ThreadTimeStamp})
		}
	default:
		logger.Warn("unsupported Events API event received")
	}
}

func (s *Slack) handleButton(evt socketmode.Event) {
	callback, ok := evt.Data.(slack.InteractionCallback)
	if !ok {
		logger.Warn("unsupported event: %+v", evt)
		return
	}
	logger.Debug("interaction received: %+v", callback)

	s.client.Ack(*evt.Request)

	switch callback.Type {
	case slack.InteractionTypeBlockActions:
		// See https://api.slack.com/apis/connections/socket-implement#button
		if len(callback.ActionCallback.BlockActions) > 0 {
			// Process wayback request from a playback action
			block := callback.ActionCallback.BlockActions[0]
			logger.Debug("received wayback action: %+v", block)
			// nolint:errcheck
			go s.process(&event{callback.User.ID, block.Value, callback.Container.ChannelID, callback.Container.MessageTs, callback.Container.ThreadTs})
		}
	case slack.InteractionTypeViewSubmission:
		// See https://api.slack.com/apis/connections/socket-implement#modal
		logger.Debug("received view submission: %+v", callback.View)
		// nolint:errcheck
		s.playback(callback.View.ExternalID, callback.View.State.Values[callbackKey][callbackKey].Value, callback.TriggerID)
	}
}

func (s *Slack) handleCommand(evt socketmode.Event) {
	cmd, ok := evt.Data.(slack.SlashCommand)
	if !ok {
		return
	}

	logger.Debug("slash command received: %+v", cmd)

	var payload interface{}
	switch cmd.Command {
	case "/help":
		payload = map[string]interface{}{
			"blocks": []slack.Block{
				slack.NewSectionBlock(
					&slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: s.opts.SlackHelptext(),
					},
					nil, nil,
				),
			}}
	case "/metrics":
		stats := metrics.Gather.Export("wayback")
		if s.opts.EnabledMetrics() && stats != "" {
			payload = map[string]interface{}{
				"blocks": []slack.Block{
					slack.NewSectionBlock(
						&slack.TextBlockObject{
							Type: slack.PlainTextType,
							Text: stats,
						},
						nil, nil,
					),
				}}
		}
	case "/playback":
		// nolint:errcheck
		s.playback(cmd.ChannelID, cmd.Text, cmd.TriggerID)
	default:
	}
	s.client.Ack(*evt.Request, payload)
}

func (s *Slack) process(ev *event) (err error) {
	content := ev.Text
	logger.Debug("content: %s", content)

	urls := service.MatchURL(s.opts, content)

	metrics.IncrementWayback(metrics.ServiceSlack, metrics.StatusRequest)
	if len(urls) == 0 {
		// nolint:errcheck
		s.reply(ev, "URL no found.")
		return errors.New("URL no found")
	}

	ev, err = s.reply(ev, "Queue...")
	if err != nil {
		logger.Error("reply queue failed: %v", err)
		return
	}
	bucket := pooling.Bucket{
		Request: func(ctx context.Context) error {
			if err := s.wayback(ctx, ev, urls); err != nil {
				logger.Error("archives failed: %v", err)
				// nolint:errcheck
				s.edit(ev.Channel, ev.ThreadTimeStamp, service.MsgWaybackRetrying)
				return err
			}
			metrics.IncrementWayback(metrics.ServiceSlack, metrics.StatusSuccess)
			return nil
		},
		Fallback: func(_ context.Context) error {
			replyText := service.MsgWaybackTimeout
			// nolint:errcheck
			s.edit(ev.Channel, ev.ThreadTimeStamp, replyText)
			metrics.IncrementWayback(metrics.ServiceSlack, metrics.StatusFailure)
			return nil
		},
	}
	s.pool.Put(bucket)

	return nil
}

func (s *Slack) wayback(ctx context.Context, ev *event, urls []*url.URL) error {
	tstamp, err := s.edit(ev.Channel, ev.ThreadTimeStamp, "Archiving...")
	if err != nil {
		logger.Error("send archiving message failed: %v", err)
		return err
	}

	do := func(cols []wayback.Collect, rdx reduxer.Reduxer) error {
		logger.Debug("reduxer: %#v", rdx)

		replyText := render.ForReply(&render.Slack{Cols: cols, Data: rdx}).String()
		logger.Debug("reply text, %s", replyText)

		if _, err := s.edit(ev.Channel, tstamp, replyText); err != nil {
			logger.Error("update message failed: %v", err)
			return err
		}

		s.pub.Spread(ctx, rdx, cols, publish.FlagSlack)

		var head = render.Title(cols, rdx)

		for _, u := range urls {
			if b, ok := rdx.Load(reduxer.Src(u.String())); ok {
				if err := service.UploadToSlack(s.bot, s.opts, b.Artifact(), ev.Channel, ev.TimeStamp, head); err != nil {
					logger.Error("upload files to slack failed: %v", err)
				}
			}
		}
		return nil
	}

	return service.Wayback(ctx, s.opts, urls, do)
}

func (s *Slack) playback(channel, text, triggerID string) error {
	logger.Debug("channel %s, playback text %s, trigger id: %s", channel, text, triggerID)
	metrics.IncrementPlayback(metrics.ServiceSlack, metrics.StatusRequest)

	urls := service.MatchURL(s.opts, text)
	if len(urls) == 0 {
		// Only the inputs in input blocks will be included in view_submission’s view.state.values: https://slack.dev/java-slack-sdk/guides/modals
		playbackHint := slack.NewTextBlockObject(slack.PlainTextType, "Playback URLs", false, false)
		playbackNameText := slack.NewTextBlockObject(slack.PlainTextType, "URLs", false, false)
		playbackPlaceholder := slack.NewTextBlockObject(slack.PlainTextType, "Please send me URLs to playback...", false, false)
		playbackNameElement := slack.NewPlainTextInputBlockElement(playbackPlaceholder, callbackKey)
		playbackNameBlock := slack.NewInputBlock(callbackKey, playbackNameText, playbackHint, playbackNameElement)
		blocks := slack.Blocks{
			BlockSet: []slack.Block{playbackNameBlock},
		}

		// TODO: l10n
		titleText := slack.NewTextBlockObject(slack.PlainTextType, "Playback", false, false)
		closeText := slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
		submitText := slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)

		modalRequest := slack.ModalViewRequest{
			Type:       slack.ViewType("modal"),
			Title:      titleText,
			Close:      closeText,
			Submit:     submitText,
			Blocks:     blocks,
			ExternalID: channel,
			// CallbackID: triggerID,
		}
		if _, err := s.bot.OpenView(triggerID, modalRequest); err != nil {
			logger.Error("error opening view: %s", err)
			return err
		}
		return nil
	}

	source := "*/playback* " + text
	msgOpts := []slack.MsgOption{
		slack.MsgOptionText(source, false),
	}
	channel, tstamp, err := s.bot.PostMessage(channel, msgOpts...)
	if err != nil {
		logger.Error("send playbak processing message failed: %v", err)
		return err
	}

	go func() {
		// nolint:errcheck
		cols, _ := wayback.Playback(s.ctx, s.opts, urls...)
		logger.Debug("playback collections: %#v", cols)

		replyText := render.ForReply(&render.Slack{Cols: cols}).String()
		if _, err := s.reply(&event{Channel: channel, TimeStamp: tstamp}, replyText); err != nil {
			metrics.IncrementPlayback(metrics.ServiceSlack, metrics.StatusFailure)
			logger.Error("send playbak results failed: %v", err)
			return
		}
		metrics.IncrementPlayback(metrics.ServiceSlack, metrics.StatusSuccess)

		// Attach a wayback button on the left
		block := slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: source,
			},
			nil,
			slack.NewAccessory(
				slack.NewButtonBlockElement(
					"",
					"original text: "+text,
					&slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: "wayback",
					},
				),
			),
		)
		msgOpts := []slack.MsgOption{
			slack.MsgOptionBlocks(block),
		}
		if _, err := s.edit(channel, tstamp, "", msgOpts...); err != nil {
			logger.Error("attach wayback button to playback text failed: %v", err)
		}
	}()
	return nil
}

func (s *Slack) reply(ev *event, text string, options ...slack.MsgOption) (*event, error) {
	if text == "" && len(options) == 0 {
		logger.Warn("text empty, skipped")
		return ev, errors.New("text empty")
	}

	msgOpts := []slack.MsgOption{
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(ev.TimeStamp), // reply as thread
		slack.MsgOptionDisableMarkdown(),
	}
	msgOpts = append(msgOpts, options...)
	_, tstamp, err := s.bot.PostMessage(ev.Channel, msgOpts...)
	if err != nil {
		logger.Error("post message failed: %v", err)
		return ev, err
	}
	// Set ThreadTimeStamp for edit message
	ev.ThreadTimeStamp = tstamp

	return ev, nil
}

func (s *Slack) edit(channel, timestamp string, text string, options ...slack.MsgOption) (string, error) {
	if text == "" && len(options) == 0 {
		logger.Warn("text empty, skipped")
		return "", errors.New("text empty")
	}

	msgOpts := []slack.MsgOption{
		slack.MsgOptionText(text, false),
		slack.MsgOptionDisableMarkdown(),
	}
	msgOpts = append(msgOpts, options...)
	_, tstamp, _, err := s.bot.UpdateMessage(channel, timestamp, msgOpts...)
	if err != nil {
		logger.Error("post message failed: %v", err)
		return "", err
	}

	return tstamp, nil
}
