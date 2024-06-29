// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.
package main

// nosemgrep: gitlab.gosec.G108-1
import (
	"log"
	"net"
	"net/http"

	"github.com/wabarc/logger"

	_ "net/http/pprof"
)

func profiling() {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		logger.Fatal("Net listen err: %v", err)
	}

	addr := listener.Addr().(*net.TCPAddr)
	logger.Info("Go profiling via: http://%s", addr)
	logger.Info("More details can be found at https://go.dev/blog/pprof")

	go func() {
		//#nosec G114 -- Ignored for convenience
		log.Println(http.Serve(listener, nil))
	}()
}
