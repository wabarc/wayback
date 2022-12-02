// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package httpd // import "github.com/wabarc/wayback/service/httpd"

import (
	"context"
	"encoding/json"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/service"
	"github.com/wabarc/wayback/template"
	"github.com/wabarc/wayback/version"
)

type web struct {
	ctx context.Context

	pool     *pooling.Pool
	router   *mux.Router
	template *template.Template
}

func newWeb(ctx context.Context, pool *pooling.Pool) *web {
	router := mux.NewRouter()
	web := &web{
		ctx:      ctx,
		pool:     pool,
		router:   router,
		template: template.New(router),
	}
	if err := web.template.ParseTemplates(); err != nil {
		logger.Fatal("unable to parse templates: %v", err)
	}
	if err := template.GenerateJavascriptBundles(); err != nil {
		logger.Fatal("unable to generate JavaScript bundles: %v", err)
	}
	return web
}

func (web *web) handle() http.Handler {
	web.router.HandleFunc("/", web.home)
	web.router.HandleFunc("/{name}.js", web.showJavascript).Name("javascript").Methods(http.MethodGet)
	web.router.HandleFunc("/favicon.ico", web.showFavicon).Name("favicon").Methods(http.MethodGet)
	web.router.HandleFunc("/icon/{filename}", web.showAppIcon).Name("icon").Methods(http.MethodGet)
	web.router.HandleFunc("/manifest.json", web.showWebManifest).Name("manifest").Methods(http.MethodGet)
	web.router.HandleFunc("/offline.html", web.showOfflinePage).Methods(http.MethodGet)

	web.router.HandleFunc("/wayback", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(web.ctx, config.Opts.WaybackTimeout())
		defer cancel()

		if err := web.process(ctx, w, r); err != nil {
			logger.Error("httpd: process retrying: %v", err)
			return
		}
		return
	}).Methods(http.MethodPost)

	web.router.HandleFunc("/playback", web.playback).Methods(http.MethodPost)

	web.router.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write(helper.String2Byte("OK")) // nolint:errcheck
	}).Name("healthcheck")

	web.router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write(helper.String2Byte(version.Version)) // nolint:errcheck
	}).Name("version")

	if config.Opts.EnabledMetrics() {
		web.router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
	}

	if config.Opts.HasDebugMode() {
		web.router.PathPrefix("/debug/").Handler(http.DefaultServeMux)
	}

	web.router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write(helper.String2Byte("User-agent: *\nDisallow: /")) // nolint:errcheck
	})

	return web.router
}

func (web *web) home(w http.ResponseWriter, r *http.Request) {
	logger.Debug("access home")
	w.Header().Set("Cache-Control", "max-age=2592000")
	if html, ok := web.template.Render("layout", nil); ok {
		w.Write(html) // nolint:errcheck
	} else {
		logger.Error("render template for home request failed")
		http.Error(w, "Internal Server Error", 500)
	}
}

func (web *web) showOfflinePage(w http.ResponseWriter, r *http.Request) {
	logger.Debug("access offline page")
	// if f, ok := w.(http.Flusher); ok {
	// 	f.Flush()
	// }
	if html, ok := web.template.Render("offline", nil); ok {
		w.Write(html) // nolint:errcheck
	} else {
		logger.Error("render template for offline request failed")
		http.Error(w, "Internal Server Error", 500)
	}
}

func (web *web) showWebManifest(w http.ResponseWriter, r *http.Request) {
	logger.Debug("access manifest")
	type webManifestIcon struct {
		Source string `json:"src"`
		Sizes  string `json:"sizes"`
		Type   string `json:"type"`
	}

	type webManifest struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		ShortName   string            `json:"short_name"`
		StartURL    string            `json:"start_url"`
		Icons       []webManifestIcon `json:"icons"`
		Display     string            `json:"display"`
		ThemeColor  string            `json:"theme_color"`
	}

	manifest := &webManifest{
		Name:        "Wayback Archiver",
		ShortName:   "Wayback",
		Description: "A toolkit for snapshot webpages",
		Display:     "standalone",
		ThemeColor:  "#f7f7f7",
		StartURL:    "/",
		Icons: []webManifestIcon{
			{Source: template.Path(web.router, "icon", "filename", "icon-120.png"), Sizes: "120x120", Type: "image/png"},
			{Source: template.Path(web.router, "icon", "filename", "icon-192.png"), Sizes: "192x192", Type: "image/png"},
			{Source: template.Path(web.router, "icon", "filename", "icon-512.png"), Sizes: "512x512", Type: "image/png"},
		},
	}

	w.Header().Set("Cache-Control", "max-age=259200")
	w.Header().Set("Content-Type", "application/manifest+json")
	if data, err := json.Marshal(manifest); err != nil {
		logger.Error("encode for response failed, %v", err)
	} else {
		w.Write(data) // nolint:errcheck
	}
}

func (web *web) showFavicon(w http.ResponseWriter, r *http.Request) {
	logger.Info("access favicon")

	blob, err := template.LoadImageFile("favicon.ico")
	if err != nil {
		return
	}
	w.Header().Set("Cache-Control", "max-age=2592000")
	w.Header().Set("Content-Type", "image/x-icon")
	w.Write(blob) // nolint:errcheck
}

func (web *web) showAppIcon(w http.ResponseWriter, r *http.Request) {
	logger.Info("access application icon")

	filename := routeParam(r, "filename")
	blob, err := template.LoadImageFile(filename)
	if err != nil {
		return
	}
	ext := path.Ext(filename)
	w.Header().Set("Cache-Control", "max-age=2592000")
	w.Header().Set("Content-Type", mime.TypeByExtension(ext))
	w.Write(blob) // nolint:errcheck
}

func (web *web) showJavascript(w http.ResponseWriter, r *http.Request) {
	filename := routeParam(r, "name")
	logger.Info("access javascript %s", filename)
	_, found := template.JavascriptBundleChecksums[filename]
	if !found {
		return
	}
	contents := template.JavascriptBundles[filename]

	w.Header().Set("Cache-Control", "max-age=2592000")
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Write(contents) // nolint:errcheck
}

func (web *web) process(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// TODO: rate limit https://pkg.go.dev/golang.org/x/time/rate
	logger.Info("process request start...")
	metrics.IncrementWayback(metrics.ServiceWeb, metrics.StatusRequest)

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusNotModified)
		return errors.New("httpd: request method no specific.")
	}

	if err := r.ParseForm(); err != nil {
		logger.Error("parse form error, %v", err)
		http.Redirect(w, r, "/", http.StatusNotModified)
		return errors.Wrap(err, "httpd: parse form error")
	}

	text := r.PostFormValue("text")
	if len(strings.TrimSpace(text)) == 0 {
		http.Redirect(w, r, "/", http.StatusFound)
		return errors.New("httpd: post form value empty")
	}
	logger.Debug("text: %s", text)

	urls := service.MatchURL(text)
	if len(urls) == 0 {
		logger.Warn("url no found.")
	}

	do := func(cols []wayback.Collect, rdx reduxer.Reduxer) error {
		collector := transform(cols)
		ctx = context.WithValue(ctx, publish.PubBundle{}, rdx)
		switch r.PostFormValue("data-type") {
		case "json":
			w.Header().Set("Content-Type", "application/json")

			if data, err := json.Marshal(collector); err != nil {
				logger.Error("encode for response failed, %v", err)
				metrics.IncrementWayback(metrics.ServiceWeb, metrics.StatusFailure)
			} else {
				if len(urls) > 0 {
					metrics.IncrementWayback(metrics.ServiceWeb, metrics.StatusSuccess)
					go publish.To(context.Background(), cols, "web")
				}
				w.Write(data) // nolint:errcheck
			}
		default:
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			if html, ok := web.template.Render("layout", collector); ok {
				if len(urls) > 0 {
					metrics.IncrementWayback(metrics.ServiceWeb, metrics.StatusSuccess)
					go publish.To(context.Background(), cols, "web")
				}
				w.Write(html) // nolint:errcheck
			} else {
				metrics.IncrementWayback(metrics.ServiceWeb, metrics.StatusFailure)
				logger.Error("render template for response failed")
			}
		}
		return nil
	}

	return service.Wayback(ctx, urls, do)
}

func (web *web) playback(w http.ResponseWriter, r *http.Request) {
	logger.Info("playback request start...")
	metrics.IncrementPlayback(metrics.ServiceWeb, metrics.StatusRequest)

	if err := r.ParseForm(); err != nil {
		logger.Error("parse form error, %v", err)
		http.Redirect(w, r, "/", http.StatusNotModified)
		return
	}

	text := r.PostFormValue("text")
	if len(strings.TrimSpace(text)) == 0 {
		logger.Warn("post form value empty.")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	logger.Debug("text: %s", text)

	urls := service.MatchURL(text)
	if len(urls) == 0 {
		logger.Warn("url no found.")
	}
	col, err := wayback.Playback(context.Background(), urls...)
	if err != nil {
		logger.Error("web: playback failed: %v", err)
		return
	}
	collector := transform(col)
	switch r.PostFormValue("data-type") {
	case "json":
		w.Header().Set("Content-Type", "application/json")

		if data, err := json.Marshal(collector); err != nil {
			logger.Error("encode for response failed, %v", err)
			metrics.IncrementPlayback(metrics.ServiceWeb, metrics.StatusFailure)
		} else {
			if len(urls) > 0 {
				metrics.IncrementPlayback(metrics.ServiceWeb, metrics.StatusSuccess)
			}
			w.Write(data) // nolint:errcheck
		}
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if html, ok := web.template.Render("layout", collector); ok {
			if len(urls) > 0 {
				metrics.IncrementPlayback(metrics.ServiceWeb, metrics.StatusSuccess)
			}
			w.Write(html) // nolint:errcheck
		} else {
			metrics.IncrementPlayback(metrics.ServiceWeb, metrics.StatusFailure)
			logger.Error("render template for response failed")
		}
	}
}

func transform(cols []wayback.Collect) template.Collector {
	collects := []template.Collect{}
	for _, col := range cols {
		collects = append(collects, template.Collect{
			Slot: col.Arc,
			Src:  col.Src,
			Dst:  col.Dst,
		})
	}
	return collects
}

func routeParam(r *http.Request, param string) string {
	vars := mux.Vars(r)
	return vars[param]
}
