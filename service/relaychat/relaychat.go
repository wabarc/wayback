// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package relaychat // import "github.com/wabarc/wayback/service/relaychat"

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gookit/color"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/ingress"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/template/render"
	"gopkg.in/irc.v4"
)

// Interface guard
var _ service.Servicer = (*IRC)(nil)

// ErrServiceClosed is returned by the Service's Serve method after a call to Shutdown.
var ErrServiceClosed = errors.New("irc: Service closed")

// IRC represents an IRC service in the application.
type IRC struct {
	ctx   context.Context
	opts  *config.Options
	pool  *pooling.Pool
	conn  *irc.Client
	store *storage.Storage
	pub   *publish.Publish
	sync.RWMutex
}

// New IRC struct.
func New(ctx context.Context, opts service.Options) (*IRC, error) {
	if !opts.Config.IRCEnabled() {
		return nil, errors.New("missing required environment variable, skipped")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	return &IRC{
		ctx:   ctx,
		store: opts.Storage,
		opts:  opts.Config,
		pool:  opts.Pool,
		pub:   opts.Publish,
	}, nil
}

// Serve loop request direct messages from the IRC server.
// Serve returns an error.
func (i *IRC) Serve() error {
	// TODO: support SASL authenticate
	config := irc.ClientConfig{
		Nick: i.opts.IRCNick(),
		User: i.opts.IRCNick(),
		Name: i.opts.IRCName(),
		Pass: i.opts.IRCPassword(),
		Handler: irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
			i.handle(c, m)
		}),
		PingFrequency: 3 * time.Second,
		PingTimeout:   time.Minute,
		SendLimit:     500 * time.Millisecond,
	}

	srv, err := url.Parse("//" + i.opts.IRCServer())
	if err != nil {
		return errors.Wrap(err, "failed to parse irc server")
	}

	secure := true
	dialer := ingress.Dialer()
conn:
	conn, err := dialer.Dial("tcp", i.opts.IRCServer())
	if err != nil {
		return errors.Wrap(err, "failed to establish connection")
	}
	if secure {
		conn = tls.Client(conn, &tls.Config{MinVersion: tls.VersionTLS12, ServerName: srv.Hostname()})
	}
	i.conn = irc.NewClient(conn, config)
	logger.Info("Serving IRC server: %s, nick: %s", i.opts.IRCServer(), color.Blue.Sprint(i.conn.CurrentNick()))

	// Block until context done
	err = i.conn.RunContext(i.ctx)
	if err != nil {
		switch {
		case err.Error() == "EOF", err.Error() == "ping timeout":
			goto conn
		case strings.HasPrefix(err.Error(), "tls:"):
			secure = false
			logger.Warn("Serving IRC server with TLS failed, fallback to non-TLS")
			goto conn
		}
		logger.Error("failed to run, error: %v", err)
		return errors.Wrap(err, "failed to run irc bot")
	}

	return ErrServiceClosed
}

// Shutdown shuts down the IRC service, it always return a nil error.
func (i *IRC) Shutdown() error {
	return nil
}

func (i *IRC) handle(c *irc.Client, m *irc.Message) {
	logger.Debug("received message: %#v", m)

	if m.Command == "ERROR" {
		logger.Error("failed to handle irc request: %s", m)
		return
	}

	if m.Command == "001" && i.opts.IRCChannel() != "" {
		if err := c.Writef("JOIN %s", i.opts.IRCChannel()); err != nil {
			logger.Error("failed to join %q channel: %v", err)
		}
		return
	}

	if m.Command == "NOTICE" && m.Name == "NickServ" && strings.Contains(m.Trailing(), "dentified") {
		logger.Debug("received command %q skipped", m.Command)
		return
	}

	if m.Command != "PRIVMSG" {
		logger.Debug("received command %q, skipped", m.Command)
		return
	}

	if c.FromChannel(m) {
		logger.Debug("received message from channel, skipped")
		return
	}

	if err := i.process(m); err != nil {
		logger.Error("process failure, message: %s, error: %v", m, err)
		return
	}
}

func (i *IRC) process(m *irc.Message) error {
	text := m.Trailing()
	urls := service.MatchURL(i.opts, text)
	logger.Debug("message body: %s", text)

	switch {
	case strings.HasPrefix(text, service.CommandHelp):
		return i.reply(m.Name, i.helper()...)

	case len(urls) == 0:
		metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusRequest)
		logger.Warn("archives failure, URL no found.")
		i.reply(m.Name, "URL no found") // nolint:errcheck

	case strings.HasPrefix(text, service.CommandPlayback):
		return i.playback(m, urls)

	case strings.HasPrefix(text, service.CommandPrivacy):
		return i.reply(m.Name, i.privacy()...)

	default:
		metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusRequest)
		i.reply(m.Name, "I'll help you archive the URL and return the results promptly.") // nolint:errcheck
		bucket := pooling.Bucket{
			Request: func(ctx context.Context) error {
				if err := i.wayback(ctx, m, urls); err != nil {
					return errors.Wrap(err, "archives failed")
				}
				metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusSuccess)
				return nil
			},
			Fallback: func(_ context.Context) error {
				i.reply(m.Name, service.MsgWaybackTimeout) // nolint:errcheck
				metrics.IncrementWayback(metrics.ServiceIRC, metrics.StatusFailure)
				return nil
			},
		}
		i.pool.Put(bucket)
	}

	return nil
}

func (i *IRC) wayback(ctx context.Context, m *irc.Message, urls []*url.URL) error {
	do := func(cols []wayback.Collect, rdx reduxer.Reduxer) error {
		logger.Debug("reduxer: %#v", rdx)

		txt := strings.Split(render.ForReply(&render.Relaychat{Cols: cols}).String(), "\n")

		ctx = context.WithValue(ctx, publish.FlagIRC, i.conn)
		i.pub.Spread(ctx, rdx, cols, publish.FlagIRC)

		return i.reply(m.Name, txt...)
	}

	return service.Wayback(ctx, i.opts, urls, do)
}

func (i *IRC) playback(m *irc.Message, urls []*url.URL) error {
	metrics.IncrementPlayback(metrics.ServiceIRC, metrics.StatusRequest)
	cols, err := wayback.Playback(i.ctx, i.opts, urls...)
	if err != nil {
		metrics.IncrementPlayback(metrics.ServiceIRC, metrics.StatusFailure)
		return errors.Wrap(err, "playback failed")
	}
	logger.Debug("playback collections: %#v", cols)

	txt := strings.Split(render.ForReply(&render.Relaychat{Cols: cols}).String(), "\n")
	if err = i.reply(m.Name, txt...); err != nil {
		logger.Error("send playback results failed: %v", err)
		return err
	}
	metrics.IncrementPlayback(metrics.ServiceTelegram, metrics.StatusSuccess)
	return nil
}

func (i *IRC) reply(name string, messages ...string) (err error) {
	for _, text := range messages {
		text = strings.ReplaceAll(text, "\n", " ")
		msg := &irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				name,
				text,
			},
		}
		if e := i.conn.WriteMessage(msg); e != nil {
			err = errors.Wrap(err, e.Error())
		}
	}
	return err
}

func (i *IRC) helper() []string {
	return []string{
		"***** List of Help *****",
		" ",
		"I'm a 🤖 to help you archive webpages more easily.",
		"Send me any text containing the URL and I'll give you the result back 😀",
		" ",
		"Examples:",
		"    /msg " + i.conn.CurrentNick() + " https://example.com",
		"    /msg " + i.conn.CurrentNick() + " playback https://example.com",
		" ",
		"Documentation:",
		"    https://docs.wabarc.eu.org",
		" ",
		"***** End of Help *****",
	}
}

func (i *IRC) privacy() []string {
	return []string{
		fmt.Sprintf("To read our privacy policy, please visit %s.", i.opts.PrivacyURL()),
	}
}
