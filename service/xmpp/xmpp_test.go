// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package xmpp // import "github.com/wabarc/wayback/service/xmpp"

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/service"
	"mellium.im/xmpp"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
)

var (
	pass  = "foobar"
	notls = "true"

	readyFeature = xmpp.StreamFeature{
		Name: xml.Name{Space: "urn:example", Local: "ready"},
		Parse: func(ctx context.Context, d *xml.Decoder, start *xml.StartElement) (bool, interface{}, error) {
			_, err := d.Token()
			return false, nil, err
		},
		Negotiate: func(ctx context.Context, session *xmpp.Session, data interface{}) (xmpp.SessionState, io.ReadWriter, error) {
			return xmpp.Ready, nil, nil
		},
	}
	negotiator = xmpp.NewNegotiator(func(*xmpp.Session, *xmpp.StreamConfig) xmpp.StreamConfig {
		return xmpp.StreamConfig{
			Features: []xmpp.StreamFeature{readyFeature},
		}
	})
	client = `<stream:stream id='316732270768047465' version='1.0' xml:lang='en' xmlns:stream='http://etherx.jabber.org/streams' xmlns='jabber:server'><stream:features><ready xmlns='urn:example'/></stream:features>`

	to   = jid.MustParse("to@example.net")
	from = jid.MustParse("from@example.net")
	msg  = stanza.Message{
		XMLName: xml.Name{Local: "message"},
		ID:      "123",
		To:      to,
		From:    from,
		Lang:    "te",
		Type:    stanza.ChatMessage,
	}
	start = msg.StartElement()
)

func setenv(t *testing.T, id string) {
	t.Setenv("WAYBACK_XMPP_JID", id)
	t.Setenv("WAYBACK_XMPP_PASSWORD", pass)
	t.Setenv("WAYBACK_XMPP_NOTLS", notls)
	t.Setenv("WAYBACK_ENABLE_IA", "true")
}

func parseOpts(t *testing.T) service.Options {
	parser := config.NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}

	pool := &pooling.Pool{}
	pub := &publish.Publish{}
	return service.ParseOptions(
		service.Config(opts),
		service.Pool(pool),
		service.Publish(pub),
	)
}

func xmppSession(ctx context.Context, t *testing.T) *xmpp.Session {
	buf := &bytes.Buffer{}
	rw := struct {
		io.Reader
		io.Writer
	}{
		Reader: strings.NewReader(client),
		Writer: buf,
	}

	session, err := xmpp.NewSession(ctx, jid.JID{}, jid.JID{}, rw, xmpp.SessionState(0), negotiator)
	if err != nil {
		t.Fatalf("Unexpected new xmpp client session: %v", err)
	}

	return session
}

func TestServe(t *testing.T) {
	opts := parseOpts(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	s := xmppSession(ctx, t)
	x := &XMPP{
		bot:  s,
		ctx:  ctx,
		opts: opts.Config,
	}
	err := x.Serve()
	if err != ErrServiceClosed {
		t.Fatalf("Unexpected serve xmpp session: %v", err)
	}
}

func TestShutdown(t *testing.T) {
	opts := parseOpts(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	s := xmppSession(ctx, t)
	x := &XMPP{
		bot:  s,
		ctx:  ctx,
		opts: opts.Config,
	}
	err := x.Shutdown()
	if err != nil {
		t.Fatalf("Unexpected shutdown xmpp session: %v", err)
	}
}

func TestProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test in short mode.")
	}

	opts := parseOpts(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	s := xmppSession(ctx, t)
	x := &XMPP{
		bot:  s,
		ctx:  ctx,
		opts: opts.Config,
		pool: opts.Pool,
	}

	mb := messageBody{msg, "foo uri"}
	err := x.process(mb)
	if err != nil {
		t.Fatalf("Error decoding message: %q", err)
	}
}

func TestWayback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test in short mode.")
	}

	opts := parseOpts(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: handle request
	})

	s := xmppSession(ctx, t)
	x := &XMPP{
		bot:  s,
		ctx:  ctx,
		opts: opts.Config,
		pool: opts.Pool,
	}

	tests := [...]struct {
		name string
		uri  string
		err  error
	}{
		{"without uri", "", fmt.Errorf("URL no found")},
		{"with uri", server.URL, context.DeadlineExceeded}, // TODO: need a complete testing
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := messageBody{msg, tt.uri}
			err := x.wayback(ctx, mb)
			if err != nil && !reflect.DeepEqual(err, tt.err) {
				t.Fatalf("Error wayback: %q", err)
			}
		})
	}
}

func TestPlayback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test in short mode.")
	}

	opts := parseOpts(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: handle request
	})

	s := xmppSession(ctx, t)
	x := &XMPP{
		bot:  s,
		ctx:  ctx,
		opts: opts.Config,
		pool: opts.Pool,
	}

	tests := [...]struct {
		name string
		uri  string
		err  error
	}{
		{"without uri", "", fmt.Errorf("URL no found")},
		// {"with uri", server.URL, context.DeadlineExceeded}, // TODO: need a complete testing
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := messageBody{msg, tt.uri}
			err := x.playback(ctx, mb)
			if err != nil && !reflect.DeepEqual(err, tt.err) {
				t.Fatalf("Error wayback: %q", err)
			}
		})
	}
}
