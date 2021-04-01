// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"bytes"
	"context"
	"crypto/tls"
	"text/template"

	irc "github.com/thoj/go-ircevent"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/logger"
)

type IRC struct {
	opts *config.Options
	conn *irc.Connection
}

func NewIRC(conn *irc.Connection, opts *config.Options) *IRC {
	if !opts.PublishToIRCChannel() {
		logger.Error("Missing required environment variable, abort.")
		return new(IRC)
	}

	if conn == nil && opts != nil {
		conn = irc.IRC(opts.IRCNick(), opts.IRCNick())
		conn.Password = opts.IRCPassword()
		conn.VerboseCallbackHandler = opts.HasDebugMode()
		conn.Debug = opts.HasDebugMode()
		conn.UseTLS = true
		conn.TLSConfig = &tls.Config{InsecureSkipVerify: false}
	}

	return &IRC{opts: opts, conn: conn}
}

func (i *IRC) ToChannel(ctx context.Context, opts *config.Options, text string) bool {
	if !opts.PublishToIRCChannel() || i.conn == nil {
		logger.Debug("[publish] Do not publish to IRC channel.")
		return false
	}

	i.conn.Join(i.opts.IRCChannel())
	i.conn.Privmsg(i.opts.IRCChannel(), text)

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

	return tmplBytes.String()
}
