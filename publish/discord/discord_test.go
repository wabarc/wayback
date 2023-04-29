// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package discord // import "github.com/wabarc/wayback/publish/discord"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"

	discord "github.com/bwmarrin/discordgo"
)

const (
	// token     = "discord-token"
	channelID = "81324113413441431"
	messageID = "100000001"
)

var (
	token    = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
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
		if err := json.Unmarshal(b, &dat); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		content := dat["content"].(string)
		uri := strings.Replace(r.URL.String(), "http:", "https:", 1)
		switch {
		case r.URL.Path == "/":
			echo(w, r)
		case r.URL.Path == "/api/v8/users/@me/channels":
			channelJson, _ := json.Marshal(channel)
			fmt.Fprintln(w, string(channelJson))
		case r.URL.Path == "/api/v8/gateway":
			gatewayJson, _ := json.Marshal(struct {
				URL string `json:"url"`
			}{URL: gateway})
			fmt.Fprintln(w, string(gatewayJson))
		case uri == discord.EndpointChannelMessages(channelID) && r.Method == http.MethodPost:
			// https://discord.com/api/v8/channels/fake-discord-channel-id/messages
			if strings.Contains(content, config.SlotName(config.SLOT_IA)) {
				fmt.Fprintln(w, `{"code":0}`)
				return
			}
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		case uri == discord.EndpointChannelMessage(channelID, messageID) && r.Method == http.MethodPatch:
			// https://discord.com/api/v8/channels/fake-discord-channel-id/messages/100000001
			if strings.Contains(content, "Archiving...") {
				fmt.Fprintln(w, string(messageJson))
				return
			}
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		default:
			fmt.Println(r.Method, r.URL.Path, content)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	})
}

func setDiscordEnv(t *testing.T) *config.Options {
	t.Setenv("WAYBACK_DISCORD_BOT_TOKEN", token)
	t.Setenv("WAYBACK_DISCORD_CHANNEL", channelID)
	t.Setenv("WAYBACK_ENABLE_IP", "true")

	opts, _ := config.NewParser().ParseEnvironmentVariables()
	return opts
}

func TestToDiscordChannel(t *testing.T) {
	opts := setDiscordEnv(t)

	httpClient, mux, server := helper.MockServer()
	defer server.Close()
	handle(mux, strings.Replace(server.URL, "http", "ws", 1))

	d := New(httpClient, opts)
	txt := render.ForPublish(&render.Discord{Cols: publish.Collects, Data: reduxer.BundleExample()}).String()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	art, err := publish.Artifact(ctx, reduxer.BundleExample(), publish.Collects)
	if err != nil {
		t.Fatalf("extract data failed: %#v", err)
	}

	got := d.toChannel(art, txt)
	if !got {
		t.Errorf("Unexpected publish to discord channel got %t instead of %t", got, true)
	}
}

func TestShutdown(t *testing.T) {
	opts := setDiscordEnv(t)

	httpClient, _, server := helper.MockServer()
	defer server.Close()

	d := New(httpClient, opts)
	err := d.Shutdown()
	if err != nil {
		t.Errorf("Unexpected shutdown: %v", err)
	}
}
