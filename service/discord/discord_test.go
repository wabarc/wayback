// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package discord // import "github.com/wabarc/wayback/service/discord"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/storage"

	discord "github.com/bwmarrin/discordgo"
)

const (
	token     = "discord-token"
	uid       = "@me"
	channelID = "81324113413441431"
	messageID = "100000001"
)

var (
	upgrader = websocket.Upgrader{}
	channel  = &discord.Channel{
		ID:   channelID,
		Name: "Discord Channel Name",
		Type: discord.ChannelTypeGuildText,
	}
	user = &discord.User{
		ID:            "-100000001",
		Email:         "bot@example.org",
		Username:      "Bot",
		Avatar:        "",
		Locale:        "en-US",
		Discriminator: "",
		Token:         "",
		Verified:      false,
		MFAEnabled:    false,
		Bot:           true,
		PublicFlags:   discord.UserFlags(1),
		PremiumType:   1,
		System:        false,
		Flags:         1,
	}
	message = discord.Message{
		ID:               messageID,
		ChannelID:        channelID,
		Content:          "https://example.com/",
		Timestamp:        discord.Timestamp("1625186466"),
		EditedTimestamp:  discord.Timestamp("1625186466"),
		MentionRoles:     []string{},
		TTS:              false,
		MentionEveryone:  false,
		Author:           user,
		Attachments:      nil,
		Components:       nil,
		Embeds:           nil,
		Mentions:         nil,
		Reactions:        nil,
		Pinned:           false,
		Type:             discord.MessageType(1),
		WebhookID:        "",
		Member:           nil,
		MentionChannels:  nil,
		Application:      nil,
		MessageReference: nil,
		Flags:            discord.MessageFlags(1),
	}
	messageJson, _ = json.Marshal(message)
)

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func handle(mux *http.ServeMux, gateway string) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		var dat map[string]interface{}
		var content string
		var once sync.Once
		if err := json.Unmarshal(b, &dat); err == nil {
			content, _ = dat["content"].(string)
		}
		uri := strings.Replace(r.URL.String(), "http:", "https:", 1)
		switch {
		case r.URL.Path == "/":
			echo(w, r)
		case uri == discord.EndpointUserChannels(uid):
			channelJson, _ := json.Marshal(channel)
			fmt.Fprintln(w, string(channelJson))
		case uri == discord.EndpointGateway:
			gatewayJson, _ := json.Marshal(struct {
				URL string `json:"url"`
			}{URL: gateway})
			fmt.Fprintln(w, string(gatewayJson))
		case r.URL.Path == "/api/v8/channels/messages":
			once.Do(func() {
				fmt.Fprintln(w, string(messageJson))
			})
		case uri == discord.EndpointChannelMessages(channelID) && r.Method == http.MethodPost:
			// https://discord.com/api/v8/channels/fake-discord-channel-id/messages
			if content == "" {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, string(messageJson))
		case uri == discord.EndpointChannelMessage(channelID, messageID) && r.Method == http.MethodPatch:
			// https://discord.com/api/v8/channels/fake-discord-channel-id/messages/100000001
			if strings.Contains(content, "Archiving...") {
				fmt.Fprintln(w, string(messageJson))
				return
			}
			if strings.Contains(content, config.SlotName(config.SLOT_IP)) {
				fmt.Fprintln(w, `{"code":0}`)
				return
			}
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		default:
			fmt.Println(uri, r.Method, r.URL.Path, content)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	})
}

func TestServe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test in short mode.")
	}

	os.Setenv("WAYBACK_DISCORD_BOT_TOKEN", token)

	parser := config.NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	opts.EnableServices(config.ServiceDiscord.String())

	httpClient, mux, server := helper.MockServer()
	defer server.Close()
	handle(mux, strings.Replace(server.URL, "http", "ws", 1))

	cfg := []pooling.Option{
		pooling.Capacity(opts.PoolingSize()),
		pooling.Timeout(opts.WaybackTimeout()),
		pooling.MaxRetries(opts.WaybackMaxRetries()),
	}
	ctx, cancel := context.WithCancel(context.Background())
	pool := pooling.New(ctx, cfg...)
	go pool.Roll()

	dbpath := filepath.Join(t.TempDir(), "testing.db")
	store, err := storage.Open(opts, dbpath)
	if err != nil {
		t.Fatalf("open storage failed: %v", err)
	}
	defer store.Close()

	pub := publish.New(ctx, opts)
	defer pub.Stop()

	o := service.ParseOptions(service.Config(opts), service.Storage(store), service.Pool(pool), service.Publish(pub))
	d, _ := New(ctx, o)
	d.bot.Client = httpClient
	time.AfterFunc(3*time.Second, func() {
		// TODO: find a better way to avoid deadlock
		go d.Shutdown()
		time.Sleep(time.Second)
		pool.Close()
		cancel()
	})
	got := d.Serve()
	expected := ErrServiceClosed
	if got != expected {
		t.Errorf("Unexpected serve telegram got %v instead of %v", got, expected)
	}
}

func TestProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test in short mode.")
	}

	os.Setenv("WAYBACK_DISCORD_BOT_TOKEN", token)
	os.Setenv("WAYBACK_DISCORD_CHANNEL", channelID)
	os.Setenv("WAYBACK_ENABLE_IP", "true")

	parser := config.NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Parse environment variables or flags failed, error: %v", err)
	}
	opts.EnableServices(config.ServiceDiscord.String())

	dbpath := filepath.Join(t.TempDir(), "testing.db")
	store, err := storage.Open(opts, dbpath)
	if err != nil {
		t.Fatalf("open storage failed: %v", err)
	}
	defer store.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cfg := []pooling.Option{
		pooling.Capacity(opts.PoolingSize()),
		pooling.Timeout(opts.WaybackTimeout()),
		pooling.MaxRetries(opts.WaybackMaxRetries()),
	}
	pool := pooling.New(ctx, cfg...)
	go pool.Roll()

	httpClient, mux, server := helper.MockServer()
	defer server.Close()
	handle(mux, strings.Replace(server.URL, "http", "ws", 1))

	pub := publish.New(ctx, opts)
	defer pub.Stop()

	o := service.ParseOptions(service.Config(opts), service.Storage(store), service.Pool(pool), service.Publish(pub))
	d, _ := New(ctx, o)
	d.bot.Client = httpClient

	// if err := d.bot.Open(); err != nil {
	// 	t.Fatal(err)
	// }
	defer d.bot.Close()

	if err := d.process(&discord.MessageCreate{Message: &message}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	pool.Close()
}
