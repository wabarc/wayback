// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"bytes"
	"context"
	"crypto/tls"
	"strings"
	"text/template"

	irc "github.com/thoj/go-ircevent"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
)

type IRC struct {
	conn *irc.Connection
}

func NewIRC(conn *irc.Connection) *IRC {
	if !config.Opts.PublishToIRCChannel() {
		logger.Error("Missing required environment variable, abort.")
		return new(IRC)
	}

	if conn == nil {
		conn = irc.IRC(config.Opts.IRCNick(), config.Opts.IRCNick())
		conn.Password = config.Opts.IRCPassword()
		conn.VerboseCallbackHandler = config.Opts.HasDebugMode()
		conn.Debug = config.Opts.HasDebugMode()
		conn.UseTLS = true
		conn.TLSConfig = &tls.Config{InsecureSkipVerify: false}
	}

	return &IRC{conn: conn}
}

func (i *IRC) ToChannel(ctx context.Context, text string) bool {
	if !config.Opts.PublishToIRCChannel() || i.conn == nil {
		logger.Debug("[publish] Do not publish to IRC channel.")
		return false
	}
	if text == "" {
		logger.Info("[publish] IRC validation failed: Text can't be blank")
		return false
	}

	i.conn.Join(config.Opts.IRCChannel())
	i.conn.Privmsg(config.Opts.IRCChannel(), text)

	return true
}

func (i *IRC) Render(vars []*wayback.Collect) string {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $ := .}}{{ $.Arc }}:- {{ range $src, $dst := $.Dst }}• {{ $dst }}, {{end}}{{end}}`

	tpl, err := template.New("message").Parse(tmpl)
	if err != nil {
		logger.Error("[publish] parse IRC template failed, %v", err)
		return ""
	}

	err = tpl.Execute(&tmplBytes, vars)
	if err != nil {
		logger.Error("[publish] execute IRC template failed, %v", err)
		return ""
	}

	return strings.TrimSuffix(tmplBytes.String(), ", ")
}
