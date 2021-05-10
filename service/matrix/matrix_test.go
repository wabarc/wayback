// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package matrix // import "github.com/wabarc/wayback/service/matrix"

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/storage"
	matrix "maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// testServer returns an http Client, ServeMux, and Server. The client proxies
// requests to the server and handlers can be registered on the mux to handle
// requests. The caller must close the test server.
func testServer() (*http.Client, *http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	transport := &RewriteTransport{&http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}}
	client := &http.Client{Transport: transport}
	return client, mux, server
}

// RewriteTransport rewrites https requests to http to avoid TLS cert issues
// during testing.
type RewriteTransport struct {
	Transport http.RoundTripper
}

// RoundTrip rewrites the request scheme to http and calls through to the
// composed RoundTripper or if it is nil, to the http.DefaultTransport.
func (t *RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	if t.Transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.Transport.RoundTrip(req)
}

func (m *Matrix) setup(roomIDs []id.RoomID) {
	clearRooms := func(m *Matrix, roomIDs []id.RoomID) {
		for _, roomID := range roomIDs {
			m.destroyRoom(roomID)
		}
	}

	if m.client != nil {
		clearRooms(m, roomIDs)
	}
}

var (
	homeserver = "https://matrix.org"
	senderUID  = os.Getenv("SENDER_UID")
	senderPwd  = os.Getenv("SENDER_PWD")
	recverUID  = os.Getenv("RECVER_UID")
	recverPwd  = os.Getenv("RECVER_PWD")
	roomID     = os.Getenv("MATRIX_ROOMID")
	err        error
	parser     *config.Parser
)

func init() {
	os.Setenv("DEBUG", "true")
	os.Setenv("WAYBACK_ENABLE_IA", "true")
	os.Setenv("WAYBACK_MATRIX_HOMESERVER", homeserver)
	os.Setenv("WAYBACK_MATRIX_ROOMID", roomID)
}

func senderClient(t *testing.T) *Matrix {
	os.Setenv("WAYBACK_MATRIX_USERID", senderUID)
	os.Setenv("WAYBACK_MATRIX_PASSWORD", senderPwd)
	parser = config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse enviroment variables or flags failed, error: %v", err)
	}
	return New(&storage.Storage{})
}

func recverClient(t *testing.T) *Matrix {
	os.Setenv("WAYBACK_MATRIX_USERID", recverUID)
	os.Setenv("WAYBACK_MATRIX_PASSWORD", recverPwd)
	parser = config.NewParser()
	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		t.Fatalf("Parse enviroment variables or flags failed, error: %v", err)
	}
	return New(&storage.Storage{})
}

func TestProcess(t *testing.T) {
	if senderUID == "" || senderPwd == "" {
		t.Skip("Define SENDER_UID and SENDER_PWD environment variables to test Matrix")
	}
	if recverUID == "" || recverPwd == "" {
		t.Skip("Define RECVER_UID and RECVER_PWD environment variables to test Matrix")
	}
	done := make(chan bool, 1)

	sender := senderClient(t)
	recver := recverClient(t)
	// sender.client.LogoutAll()
	// recver.client.LogoutAll()
	sender.setup(sender.joinedRooms())
	recver.setup(recver.joinedRooms())

	// Mock Client
	httpClient, mux, server := testServer()
	defer server.Close()

	// TODO: mock
	// see https://matrix.org/docs/spec/client_server/latest#post-matrix-client-r0-createroom
	mux.HandleFunc("/_matrix/client/r0/createRoom", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"room_id": "!sefiuhWgwghwWgh:example.com"}`)
	})
	mux.HandleFunc("/_matrix/client/r0/rooms/!sefiuhWgwghwWgh:example.com/send/m.room.message/mautrix-go_1617716651413791400_1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"room_id": "!sefiuhWgwghwWgh:example.com"}`)
	})
	t.Log(httpClient)
	// sender.client.Client = httpClient

	// Create a room and invite recver
	resp, err := sender.client.CreateRoom(&matrix.ReqCreateRoom{
		Invite:     []id.UserID{id.UserID(config.Opts.MatrixUserID())},
		Preset:     "trusted_private_chat",
		Visibility: "private",
		IsDirect:   true,
	})
	if err != nil {
		t.Fatalf("Create room failure, error: %v", err)
	}

	// Send message to recver
	if _, err = sender.client.SendText(resp.RoomID, "Hello, https://example.com?r="+helper.RandString(3, "")); err != nil {
		t.Fatalf("Send text to recver failure, error: %v", err)
	}

	// Listen message event from sender
	recvSyncer := recver.client.Syncer.(*matrix.DefaultSyncer)
	recvSyncer.OnEventType(event.StateMember, func(source matrix.EventSource, ev *event.Event) {
		ms := ev.Content.AsMember().Membership
		if ev.Sender == id.UserID(senderUID) && ms == event.MembershipInvite {
			t.Logf("Event id: %s, event type: %s, event content: %v", id.EventID(ev.ID), ev.Type.Type, ev.Content.Raw)
			if _, err := recver.client.JoinRoomByID(ev.RoomID); err != nil {
				t.Fatalf("Accept invitation from sender failure, error: %v", err)
			}
		}
	})
	recvSyncer.OnEventType(event.EventMessage, func(source matrix.EventSource, ev *event.Event) {
		if ev.Sender == id.UserID(senderUID) {
			t.Logf("Event id: %s, event type: %s, event content: %v", id.EventID(ev.ID), ev.Type.Type, ev.Content.AsMessage().Body)

			if err := recver.process(context.Background(), ev); err != nil {
				t.Errorf("Process request failure, error: %v", err)
			}
			done <- true
		}
	})
	recvSyncer.OnEventType(event.EventEncrypted, func(source matrix.EventSource, ev *event.Event) {
		t.Log("Unsupport encryption message")
		// logger.Debug("[matrix] event: %v", ev)
		// if err := m.process(context.Background(), ev); err != nil {
		// 	logger.Error("[matrix] process request failure, error: %v", err)
		// }
	})

	go func() {
		tick := time.NewTicker(time.Second)
		i := 60
		for {
			select {
			case <-tick.C:
				if i == 0 {
					t.Error("Timeout while wating for test message from the other thread.")
					sender.destroyRoomForTest(resp.RoomID)
					recver.destroyRoomForTest(resp.RoomID)
					time.Sleep(time.Second)
					recver.client.StopSync()
					sender.client.StopSync()
					return
				}
			case <-done:
				tick.Stop()
				sender.destroyRoomForTest(resp.RoomID)
				recver.destroyRoomForTest(resp.RoomID)
				time.Sleep(time.Second)
				recver.client.StopSync()
				sender.client.StopSync()
			}
			i -= 1
		}
	}()

	go func() {
		for {
			if err := recver.client.Sync(); err != nil {
				t.Log(err)
			}
		}
	}()
	if err := sender.client.Sync(); err != nil {
		t.Log(err)
	}
}

func (m *Matrix) destroyRoomForTest(roomID id.RoomID) {
	if roomID == "" || m == nil {
		return
	}

	m.client.LeaveRoom(roomID)
	m.client.ForgetRoom(roomID)
}
