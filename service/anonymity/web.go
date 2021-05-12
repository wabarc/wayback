// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package anonymity // import "github.com/wabarc/wayback/service/anonymity"

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wabarc/helper"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/template"
	"github.com/wabarc/wayback/version"
)

type web struct {
	router   *mux.Router
	template *template.Template
}

func newWeb() *web {
	router := mux.NewRouter()
	web := &web{
		router:   router,
		template: template.New(router),
	}
	if err := web.template.ParseTemplates(); err != nil {
		logger.Fatal("[web] unable to parse templates: %v", err)
	}
	if err := template.GenerateJavascriptBundles(); err != nil {
		logger.Fatal("[web] unable to generate JavaScript bundles: %v", err)
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

	web.router.HandleFunc("/w", func(w http.ResponseWriter, r *http.Request) { web.process(w, r) }).Methods(http.MethodPost)

	web.router.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}).Name("healthcheck")
	web.router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(version.Version))
	}).Name("version")
	if config.Opts.EnabledMetrics() {
		web.router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
	}

	web.router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("User-agent: *\nDisallow: /"))
	})

	return web.router
}

func (web *web) home(w http.ResponseWriter, r *http.Request) {
	logger.Debug("[web] access home")
	w.Header().Set("Cache-Control", "max-age=2592000")
	if html, ok := web.template.Render("layout", nil); ok {
		w.Write(html)
	} else {
		logger.Error("[web] render template for home request failed")
		http.Error(w, "Internal Server Error", 500)
	}
}

func (web *web) showOfflinePage(w http.ResponseWriter, r *http.Request) {
	logger.Debug("[web] access offline page")
	// if f, ok := w.(http.Flusher); ok {
	// 	f.Flush()
	// }
	if html, ok := web.template.Render("offline", nil); ok {
		w.Write(html)
	} else {
		logger.Error("[web] render template for offline request failed")
		http.Error(w, "Internal Server Error", 500)
	}
}

func (web *web) showWebManifest(w http.ResponseWriter, r *http.Request) {
	logger.Debug("[web] access manifest")
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
		logger.Error("[web] encode for response failed, %v", err)
	} else {
		w.Write(data)
	}
}

func (web *web) showFavicon(w http.ResponseWriter, r *http.Request) {
	logger.Debug("[web] access favicon")

	blob, err := template.LoadImageFile("favicon.ico")
	if err != nil {
		return
	}
	w.Header().Set("Cache-Control", "max-age=2592000")
	w.Header().Set("Content-Type", "image/x-icon")
	w.Write(blob)
}

func (web *web) showAppIcon(w http.ResponseWriter, r *http.Request) {
	logger.Debug("[web] access application icon")

	filename := routeParam(r, "filename")
	blob, err := template.LoadImageFile(filename)
	if err != nil {
		return
	}
	w.Header().Set("Cache-Control", "max-age=2592000")
	w.Header().Set("Content-Type", "image/png")
	w.Write(blob)
}

func (web *web) showJavascript(w http.ResponseWriter, r *http.Request) {
	filename := routeParam(r, "name")
	logger.Debug("[web] access javascript %s", filename)
	_, found := template.JavascriptBundleChecksums[filename]
	if !found {
		return
	}
	contents := template.JavascriptBundles[filename]

	w.Header().Set("Cache-Control", "max-age=2592000")
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Write(contents)
}

func (web *web) process(w http.ResponseWriter, r *http.Request) {
	logger.Debug("[web] process request start...")
	metrics.IncrementWayback(metrics.ServiceWeb, metrics.StatusRequest)

	if r.Method != http.MethodPost {
		logger.Info("[web] request method no specific.")
		http.Redirect(w, r, "/", http.StatusNotModified)
		return
	}

	if err := r.ParseForm(); err != nil {
		logger.Error("[web] parse form error, %v", err)
		http.Redirect(w, r, "/", http.StatusNotModified)
		return
	}

	text := r.PostFormValue("text")
	if len(strings.TrimSpace(text)) == 0 {
		logger.Info("[web] post form value empty.")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	logger.Debug("[web] text: %s", text)

	urls := helper.MatchURLFallback(text)
	if len(urls) == 0 {
		logger.Info("[web] url no found.")
	}
	col, _ := wayback.Wayback(urls)
	collector := transform(col)
	ctx := context.Background()
	switch r.PostFormValue("data-type") {
	case "json":
		w.Header().Set("Content-Type", "application/json")

		if data, err := json.Marshal(collector); err != nil {
			logger.Error("[web] encode for response failed, %v", err)
			metrics.IncrementWayback(metrics.ServiceWeb, metrics.StatusFailure)
		} else {
			metrics.IncrementWayback(metrics.ServiceWeb, metrics.StatusSuccess)
			if len(urls) > 0 {
				go publish.To(ctx, col, "web")
			}
			w.Write(data)
		}
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if html, ok := web.template.Render("layout", collector); ok {
			metrics.IncrementWayback(metrics.ServiceWeb, metrics.StatusSuccess)
			if len(urls) > 0 {
				go publish.To(ctx, col, "web")
			}
			w.Write(html)
		} else {
			metrics.IncrementWayback(metrics.ServiceWeb, metrics.StatusFailure)
			logger.Error("[web] render template for response failed")
		}
	}
}

func transform(col []*wayback.Collect) template.Collector {
	collects := []template.Collect{}
	for _, c := range col {
		for src, dst := range c.Dst {
			collects = append(collects, template.Collect{
				Slot: c.Arc,
				Src:  src,
				Dst:  dst,
			})
		}
	}
	return collects
}

func routeParam(r *http.Request, param string) string {
	vars := mux.Vars(r)
	return vars[param]
}
