// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package relaychat // import "github.com/wabarc/wayback/service/relaychat"

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"
	"gopkg.in/irc.v4"
)

var (
	server   = "irc.libera.chat:6697"
	sender   = "wsend" + helper.RandString(4, "lower")
	receiver = "wrecv" + helper.RandString(4, "lower")
	channel  = "#wabarc-testing"
	debug    = false
)

type TestHandler struct {
	messages []*irc.Message
	delay    time.Duration
}

func (th *TestHandler) Handle(c *irc.Client, m *irc.Message) {
	th.messages = append(th.messages, m)
	if th.delay > 0 {
		time.Sleep(th.delay)
	}
}

func (th *TestHandler) Messages() []*irc.Message {
	ret := th.messages
	th.messages = nil
	return ret
}

var errorWriterErr = errors.New("errorWriter: error")

type errorWriter struct{}

func (ew *errorWriter) Write([]byte) (int, error) {
	return 0, errorWriterErr
}

type readWriteCloser struct {
	io.Reader
	io.Writer
	io.Closer
}

type testReadWriteCloser struct {
	client *bytes.Buffer
	server *bytes.Buffer
}

type testReadWriter struct {
	writeErrorChan chan error
	writeChan      chan string
	readErrorChan  chan error
	readChan       chan string
	readEmptyChan  chan struct{}
	exiting        chan struct{}
	clientDone     chan struct{}
	closed         bool
	serverBuffer   bytes.Buffer
}

func (rw *testReadWriter) maybeBroadcastEmpty() {
	if rw.serverBuffer.Len() == 0 {
		select {
		case rw.readEmptyChan <- struct{}{}:
		default:
		}
	}
}

func (rw *testReadWriter) Read(buf []byte) (int, error) {
	// Check for a read error first
	select {
	case err := <-rw.readErrorChan:
		return 0, err
	default:
	}

	// If there's data left in the buffer, we want to use that first.
	if rw.serverBuffer.Len() > 0 {
		s, err := rw.serverBuffer.Read(buf)
		if errors.Is(err, io.EOF) {
			err = nil
		}
		rw.maybeBroadcastEmpty()
		return s, err
	}

	// Read from server. We're waiting for this whole test to finish, data to
	// come in from the server buffer, or for an error. We expect only one read
	// to be happening at once.
	select {
	case err := <-rw.readErrorChan:
		return 0, err
	case data := <-rw.readChan:
		rw.serverBuffer.WriteString(data)
		s, err := rw.serverBuffer.Read(buf)
		if errors.Is(err, io.EOF) {
			err = nil
		}
		rw.maybeBroadcastEmpty()
		return s, err
	case <-rw.exiting:
		return 0, io.EOF
	}
}

func (rw *testReadWriter) Write(buf []byte) (int, error) {
	select {
	case err := <-rw.writeErrorChan:
		return 0, err
	default:
	}

	// Write to server. We can cheat with this because we know things
	// will be written a line at a time.
	select {
	default:
		return len(buf), nil
	case rw.writeChan <- string(buf):
		return len(buf), nil
	case <-rw.exiting:
		return 0, errors.New("Connection closed")
	}
}

func (rw *testReadWriter) Close() error {
	select {
	case <-rw.exiting:
		return errors.New("Connection closed")
	default:
		// Ensure no double close
		if !rw.closed {
			rw.closed = true
			close(rw.exiting)
		}
		return nil
	}
}

func newTestReadWriter() *testReadWriter {
	return &testReadWriter{
		writeErrorChan: make(chan error, 1),
		writeChan:      make(chan string),
		readErrorChan:  make(chan error, 1),
		readChan:       make(chan string),
		readEmptyChan:  make(chan struct{}, 1),
		exiting:        make(chan struct{}),
		clientDone:     make(chan struct{}),
	}
}

func runClientTest(
	t *testing.T,
	cc irc.ClientConfig,
	expectedErr error,
	setup func(c *irc.Client),
) *irc.Client {
	t.Helper()

	rw := newTestReadWriter()
	c := irc.NewClient(rw, cc)

	if setup != nil {
		setup(c)
	}

	go func(t *testing.T) {
		err := c.Run()
		if !reflect.DeepEqual(expectedErr, err) {
			// t.Fatalf("unexpected error, got %v instead of %v", err, expectedErr)
		}
		close(rw.clientDone)
	}(t)

	runTest(t, rw)

	return c
}

func runTest(t *testing.T, rw *testReadWriter) {
	t.Helper()

	// Ask everything to shut down
	rw.Close()

	// Wait for the client to stop
	select {
	case <-rw.clientDone:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout in client shutdown")
	}
}

// Bash: echo -e 'USER wabarc-sender guest * *\nNICK wabarc-sender\nPRIVMSG wabarc-receiver :Hello, World!\nQUIT\n' \ | nc irc.freenode.net 6667
func TestProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test in short mode.")
	}

	os.Setenv("WAYBACK_IRC_NICK", "wabarc-process")
	os.Setenv("WAYBACK_IRC_SERVER", server)
	os.Setenv("WAYBACK_IRC_CHANNEL", channel)
	os.Setenv("WAYBACK_ENABLE_IA", "true")

	parser := config.NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	opts.EnableServices(config.ServiceIRC.String())

	cfg := []pooling.Option{
		pooling.Capacity(opts.PoolingSize()),
		pooling.Timeout(opts.WaybackTimeout()),
		pooling.MaxRetries(opts.WaybackMaxRetries()),
	}
	ctx := context.Background()
	pool := pooling.New(ctx, cfg...)
	go pool.Roll()
	defer pool.Close()

	pub := publish.New(ctx, opts)
	defer pub.Stop()

	cc := irc.ClientConfig{
		Nick: opts.IRCNick(),
		User: opts.IRCNick(),
		Name: opts.IRCName(),
		Pass: opts.IRCPassword(),
	}
	rw := newTestReadWriter()
	c := irc.NewClient(rw, cc)

	o := service.ParseOptions(service.Config(opts), service.Storage(&storage.Storage{}), service.Pool(pool), service.Publish(pub))
	i, err := New(context.Background(), o)
	if err != nil {
		t.Fatalf("unexpected to new an irc client: %v", err)
	}
	i.conn = c
	defer i.Shutdown()

	tests := []struct {
		desc string
		text string
		want error
	}{
		{"test help command", "help", nil},
		{"test wayback without url", ":foo bar", nil},
		{"test wayback with url", ":foo bar https://example.com", nil},
		{"test playback without url", ":playback", nil},
		{"test playback with url", ":playback https://example.com", nil},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := irc.MustParseMessage("privmsg " + tt.text)
			err = i.process(m)
			if !reflect.DeepEqual(err, tt.want) {
				// TODO: assert error
				t.Fatal(err)
			}
		})
	}
}
