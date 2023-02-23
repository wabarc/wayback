// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/template/render"
	"golang.org/x/net/websocket"
)

func TestToNostr(t *testing.T) {
	unsetAllEnv()

	// test note to be sent over websocket
	sk, _ := makeKeyPair(t)
	nsec, err := nip19.EncodePrivateKey(sk)
	if err != nil {
		t.Fatalf("encode private key err: %v", err)
	}

	// fake relay server
	var mu sync.Mutex // guards published to satisfy go test -race
	var published bool
	ws := newWebsocketServer(func(conn *websocket.Conn) {
		mu.Lock()
		published = true
		mu.Unlock()
		// verify the client sent exactly the textNote
		var raw []json.RawMessage
		if err := websocket.JSON.Receive(conn, &raw); err != nil {
			t.Fatalf("websocket.JSON.Receive: %v", err)
		}
		event := parseEventMessage(t, raw)
		s := "IPFS"
		if !bytes.ContainsAny(event.Serialize(), s) {
			t.Fatalf("received event:\n%+v\nwant:\n%+v", event, s)
		}
		// send back an ok nip-20 command result
		res := []any{"OK", event.ID, true, ""}
		if err := websocket.JSON.Send(conn, res); err != nil {
			t.Fatalf("websocket.JSON.Send: %v", err)
		}
	})
	defer ws.Close()

	os.Setenv("WAYBACK_NOSTR_RELAY_URL", ws.URL)
	os.Setenv("WAYBACK_NOSTR_PRIVATE_KEY", nsec)
	opts, _ := config.NewParser().ParseEnvironmentVariables()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	nos := NewNostr(nil, opts)
	txt := render.ForPublish(&render.Nostr{Cols: collects}).String()
	err = nos.publish(ctx, txt)
	if err != nil {
		t.Errorf("Unexpected publish nostr note got err: %v", err)
	}

	if !published {
		t.Errorf("fake relay server saw no event")
	}
}

func newWebsocketServer(handler func(*websocket.Conn)) *httptest.Server {
	return httptest.NewServer(&websocket.Server{
		Handshake: anyOriginHandshake,
		Handler:   handler,
	})
}

// anyOriginHandshake is an alternative to default in golang.org/x/net/websocket
// which checks for origin. nostr client sends no origin and it makes no difference
// for the tests here anyway.
var anyOriginHandshake = func(conf *websocket.Config, r *http.Request) error {
	return nil
}

func makeKeyPair(t *testing.T) (priv, pub string) {
	t.Helper()
	privkey := nostr.GeneratePrivateKey()
	pubkey, err := nostr.GetPublicKey(privkey)
	if err != nil {
		t.Fatalf("GetPublicKey(%q): %v", privkey, err)
	}
	return privkey, pubkey
}

func parseEventMessage(t *testing.T, raw []json.RawMessage) nostr.Event {
	t.Helper()
	if len(raw) < 2 {
		t.Fatalf("len(raw) = %d; want at least 2", len(raw))
	}
	var typ string
	json.Unmarshal(raw[0], &typ)
	if typ != "EVENT" {
		t.Errorf("typ = %q; want EVENT", typ)
	}
	var event nostr.Event
	if err := json.Unmarshal(raw[1], &event); err != nil {
		t.Errorf("json.Unmarshal: %v", err)
	}
	return event
}
