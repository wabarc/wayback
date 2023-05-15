// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package ingress // import "github.com/wabarc/wayback/ingress"

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/wabarc/logger"
	"github.com/wabarc/proxier"
	"github.com/wabarc/wayback/config"
)

var (
	client   = &http.Client{}
	endpoint = "https://icanhazip.com"
)

func initClient(opts *config.Options) {
	if opts.Proxy() != "" {
		u, err := url.Parse(opts.Proxy())
		if err != nil {
			logger.Error("proxy format invalid: %v", err)
			return
		}
		if !canConnect(u.Hostname(), u.Port()) {
			logger.Warn("proxy %s can't connect", u)
			return
		}

		client.Transport, err = proxier.NewUTLSRoundTripper(proxier.Proxy(opts.Proxy()))
		if err != nil {
			logger.Error("create utls round tripper failed: %v", err)
		}
		if opts.HasDebugMode() {
			go func() {
				for {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil) // nolint:errcheck
					resp, err := client.Do(req)
					if err != nil {
						logger.Error("request error: %v", err)
						cancel()
						continue
					}
					cancel()
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						logger.Error("read body error: %v", err)
						continue
					}
					logger.Debug("client handshake: %s", bytes.TrimSpace(body))
					resp.Body.Close()
					time.Sleep(time.Minute)
				}
			}()
		}
	}
}

func canConnect(host, port string) bool {
	addr := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return false
	}
	if conn != nil {
		return true
	}
	return false
}

// Client returns http.Client
func Client() *http.Client {
	return client
}