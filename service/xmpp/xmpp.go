// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package xmpp // import "github.com/wabarc/wayback/service/xmpp"

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

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
	"mellium.im/sasl"
	"mellium.im/xmlstream"
	"mellium.im/xmpp"
	"mellium.im/xmpp/dial"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
)

// Interface guard
var _ service.Servicer = (*XMPP)(nil)

// ErrServiceClosed is returned by the Service's Serve method after a call to Shutdown.
var ErrServiceClosed = errors.New("xmpp: Service closed")

// XMPP represents an XMPP service in the application.
type XMPP struct {
	ctx   context.Context
	bot   *xmpp.Session
	opts  *config.Options
	pool  *pooling.Pool
	store *storage.Storage
	pub   *publish.Publish
}

// messageBody is a message stanza that contains a body. It is normally used for
// chat messages.
type messageBody struct {
	stanza.Message
	Body string `xml:"body"`
}

// New XMPP struct.
func New(ctx context.Context, opts service.Options) (*XMPP, error) {
	if !opts.Config.XMPPEnabled() {
		return nil, errors.New("missing required environment variable, skipped")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	// Parse and set up JID.
	id, err := jid.Parse(opts.Config.XMPPUsername())
	if err != nil {
		return nil, errors.Wrap(err, "parsing JID failed")
	}

	// Enable optional features and initialize client session, according to configuration.
	features := []xmpp.StreamFeature{xmpp.BindResource()}
	if opts.Config.XMPPNoTLS() {
		features = append(features, xmpp.StartTLS(&tls.Config{
			ServerName: id.Domain().String(),
			MinVersion: tls.VersionTLS12,
		}))
	}
	var defaultAuthMechanisms = []sasl.Mechanism{
		sasl.Plain,
		sasl.ScramSha1,
		sasl.ScramSha1Plus,
	}

	if opts.Config.XMPPPassword() != "" {
		features = append(features, xmpp.SASL("", opts.Config.XMPPPassword(), defaultAuthMechanisms...))
	}

	dialCtx, dialCtxCancel := context.WithTimeout(ctx, 30*time.Second)
	defer dialCtxCancel()

	// Initialze connection according to configuration.
	dialer := &dial.Dialer{NoTLS: opts.Config.XMPPNoTLS(), NoLookup: opts.Config.XMPPNoTLS()}
	conn, err := dialer.Dial(dialCtx, "tcp", id)
	if err != nil {
		return nil, errors.Wrap(err, "establishing connection failed")
	}

	bot, err := xmpp.NewClientSession(dialCtx, id, conn, features...)
	if err != nil {
		return nil, errors.Wrap(err, "new xmpp client failed")
	}

	// Send initial presence to let the server know we want to receive messages.
	err = bot.Send(ctx, stanza.Presence{Type: stanza.AvailablePresence}.Wrap(nil))
	if err != nil {
		return nil, errors.Wrap(err, "error sending initial presence")
	}

	return &XMPP{
		ctx:   ctx,
		bot:   bot,
		store: opts.Storage,
		opts:  opts.Config,
		pool:  opts.Pool,
		pub:   opts.Publish,
	}, nil
}

// Serve loop request direct messages from the XMPP server.
// Serve returns an error.
func (x *XMPP) Serve() error {
	addr := x.bot.LocalAddr()
	logger.Info("Serving XMPP JID: %s", addr)

	// Handle incoming messages.
	go func() {
		err := x.bot.Serve(xmpp.HandlerFunc(func(t xmlstream.TokenReadEncoder, start *xml.StartElement) error {
			// This is a workaround for https://mellium.im/issue/196
			// until a cleaner permanent fix is devised (see https://mellium/issue/197)
			d := xml.NewTokenDecoder(xmlstream.MultiReader(xmlstream.Token(*start), t))
			if _, err := d.Token(); err != nil {
				return err
			}

			// Ignore anything that's not a message. In a real system we'd want to at
			// least respond to IQs.
			if start.Name.Local != "message" {
				return nil
			}

			msg := messageBody{}
			err := d.DecodeElement(&msg, start)
			if err != nil && err != io.EOF {
				logger.Error("Error decoding message: %q", err)
				return nil
			}

			// Don't reflect messages unless they are chat messages and actually have a
			// body.
			// In a real world situation we'd probably want to respond to IQs, at least.
			if msg.Body == "" || msg.Type != stanza.ChatMessage {
				return nil
			}

			err = x.process(msg)
			if err != nil {
				logger.Error("process failed: %v", err)
			}

			return err
		}))
		if err != nil {
			logger.Error("serve xmpp error: %v", err)
		}
	}()

	// Block until context cone
	<-x.ctx.Done()

	return ErrServiceClosed
}

// Shutdown shuts down the XMPP service, it always retuan a nil error.
func (x *XMPP) Shutdown() error {
	if err := x.bot.Close(); err != nil {
		return err
	}

	if err := x.bot.Conn().Close(); err != nil {
		return err
	}

	return nil
}

func (x *XMPP) process(msg messageBody) error {
	cmdctx, cancel := context.WithTimeout(x.ctx, 5*time.Second)
	defer cancel()

	cmd := command(msg)
	switch cmd {
	case service.CommandHelp:
		return x.reply(cmdctx, msg, x.opts.XMPPHelptext())
	case service.CommandMetrics:
		stats := metrics.Gather.Export("wayback")
		if x.opts.EnabledMetrics() && stats != "" {
			if err := x.reply(cmdctx, msg, stats); err != nil {
				return err
			}
		}
		return nil
	case service.CommandPlayback:
		return x.playback(cmdctx, msg)
	case service.CommandPrivacy:
		return x.reply(cmdctx, msg, fmt.Sprintf("To read our privacy policy, please visit %s.", x.opts.PrivacyURL()))
	default:
		metrics.IncrementWayback(metrics.ServiceXMPP, metrics.StatusRequest)
		bucket := pooling.Bucket{
			Request: func(ctx context.Context) error {
				if err := x.wayback(ctx, msg); err != nil {
					logger.Error("process failure, message: %s, error: %v", msg.Body, err)
					return err
				}
				metrics.IncrementWayback(metrics.ServiceXMPP, metrics.StatusSuccess)
				return nil
			},
			Fallback: func(ctx context.Context) error {
				if err := x.reply(ctx, msg, service.MsgWaybackTimeout); err != nil {
					logger.Error("process failure: %v", err)
				}
				metrics.IncrementWayback(metrics.ServiceXMPP, metrics.StatusFailure)
				return nil
			},
		}
		x.pool.Put(bucket)
	}
	return nil
}

func (x *XMPP) wayback(ctx context.Context, msg messageBody) error {
	text := msg.Body
	logger.Debug("received message: %s", text)

	urls := service.MatchURL(x.opts, text)
	if len(urls) == 0 {
		return x.reply(ctx, msg, "URL no found")
	}

	do := func(cols []wayback.Collect, rdx reduxer.Reduxer) error {
		logger.Debug("reduxer: %#v", rdx)

		text := render.ForReply(&render.XMPP{Cols: cols}).String()
		err := x.reply(ctx, msg, text)
		if err != nil {
			return err
		}

		x.pub.Spread(ctx, rdx, cols, publish.FlagXMPP)

		return nil
	}

	return service.Wayback(ctx, x.opts, urls, do)
}

func (x *XMPP) playback(ctx context.Context, msg messageBody) error {
	metrics.IncrementPlayback(metrics.ServiceXMPP, metrics.StatusRequest)

	urls := service.MatchURL(x.opts, msg.Body)
	if len(urls) == 0 {
		return x.reply(ctx, msg, "URL no found")
	}

	cols, err := wayback.Playback(ctx, x.opts, urls...)
	if err != nil {
		metrics.IncrementPlayback(metrics.ServiceXMPP, metrics.StatusFailure)
		return errors.Wrap(err, "xmpp: playback failed")
	}
	logger.Debug("playback collections: %#v", cols)

	text := render.ForReply(&render.XMPP{Cols: cols}).String()
	if err := x.reply(ctx, msg, text); err != nil {
		metrics.IncrementPlayback(metrics.ServiceXMPP, metrics.StatusFailure)
		logger.Error("send playback results failed: %v", err)
		return err
	}
	metrics.IncrementPlayback(metrics.ServiceXMPP, metrics.StatusSuccess)
	return nil
}

func (x *XMPP) reply(ctx context.Context, msg messageBody, s string) error {
	message := stanza.Message{
		To:   msg.From,
		Type: msg.Type,
	}
	body := messageBody{
		Message: message,
		Body:    fmt.Sprintf("%s\n%s", quote(msg.Body), s),
	}

	if err := x.bot.Encode(ctx, body); err != nil {
		return err
	}

	return nil
}

func quote(s string) string {
	sb := strings.Builder{}
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		_, _ = sb.WriteString("> " + scanner.Text() + "\n")
	}
	if err := scanner.Err(); err != nil {
		return s
	}
	return sb.String()
}

func command(msg messageBody) string {
	body := strings.TrimSpace(msg.Body)
	switch {
	case strings.HasPrefix(body, service.CommandHelp),
		strings.HasPrefix(body, "/"+service.CommandHelp),
		strings.HasPrefix(body, service.CommandHelp+":"):
		return service.CommandHelp
	case strings.HasPrefix(body, service.CommandMetrics),
		strings.HasPrefix(body, "/"+service.CommandMetrics),
		strings.HasPrefix(body, service.CommandMetrics+":"):
		return service.CommandMetrics
	case strings.HasPrefix(body, service.CommandPlayback),
		strings.HasPrefix(body, "/"+service.CommandPlayback),
		strings.HasPrefix(body, service.CommandPlayback+":"):
		return service.CommandPlayback
	case strings.HasPrefix(body, service.CommandPrivacy),
		strings.HasPrefix(body, "/"+service.CommandPrivacy),
		strings.HasPrefix(body, service.CommandPrivacy+":"):
		return service.CommandPrivacy
	}
	return "unknown"
}
