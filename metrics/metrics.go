// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package metrics // import "github.com/wabarc/wayback/metrics"

import (
	"fmt"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/version"
)

// Common status values
const (
	ServiceIRC      = "irc"
	ServiceWeb      = "web"
	ServiceSlack    = "slack"
	ServiceMatrix   = "matrix"
	ServiceDiscord  = "discord"
	ServiceMastodon = "mastodon"
	ServiceTelegram = "telegram"
	ServiceTwitter  = "twitter"

	PublishIRC     = "irc"      // IRC channel
	PublishChannel = "channel"  // Telegram channel
	PublishMstdn   = "mastodon" // Mastodon toot
	PublishGithub  = "github"   // GitHub issues
	PublishMatrix  = "room"
	PublishSlack   = "slack"
	PublishDiscord = "discord" // Discord channel
	PublishTwitter = "tweet"

	StatusRequest = "request"
	StatusSuccess = "success"
	StatusFailure = "failure"
)

// Prometheus Metrics
var (
	waybackGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "wayback",
		Name:      "wayback",
		Help:      "Total number of wayback requests from configured services",
	}, []string{"from", "status"})

	playbackGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "wayback",
		Name:      "playback",
		Help:      "Total number of playback requests from configured services",
	}, []string{"from", "status"})

	publishGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "wayback",
		Name:      "publish",
		Help:      "Total number of wayback results published to configured services",
	}, []string{"desc", "status"})

	buildInfoGauge = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "wayback",
		Name:      "info",
		Help:      "The build information of wayback service",
		ConstLabels: prometheus.Labels{
			"version":   version.Version,
			"commit":    version.Commit,
			"goversion": runtime.Version(),
			"buildDate": version.BuildDate,
			"platform":  runtime.GOOS + "/" + runtime.GOARCH,
		},
	}, func() float64 { return 1 })
)

// IncrementWayback increments the incoming requests counter
func IncrementWayback(from, status string) {
	waybackGauge.With(prometheus.Labels{"from": from, "status": status}).Inc()
}

// IncrementPlayback increments the incoming requests counter
func IncrementPlayback(from, status string) {
	playbackGauge.With(prometheus.Labels{"from": from, "status": status}).Inc()
}

// IncrementPublish increments the publish counter
func IncrementPublish(desc, status string) {
	publishGauge.With(prometheus.Labels{"desc": desc, "status": status}).Inc()
}

// Collector represents a metric collector.
type Collector struct {
	// WaybackPgs reports the archiving result for configured services
	WaybackPgs prometheus.GaugeVec

	// PlaybackPgs reports the playback result for configured services
	PlaybackPgs prometheus.GaugeVec

	// PublishPgs reports the publish result for configured services
	PublishPgs prometheus.GaugeVec

	// uptimeDesc reports the uptime of the wayback
	uptimeDesc *prometheus.Desc
}

// NewCollector initializes a new metric collector.
func NewCollector() *Collector {
	collector := &Collector{
		WaybackPgs:  *waybackGauge,
		PlaybackPgs: *playbackGauge,
		PublishPgs:  *publishGauge,
		uptimeDesc: prometheus.NewDesc(
			"wayback_uptime",
			"The uptime of wayback service.",
			[]string{"duration"}, nil),
	}
	prometheus.MustRegister(collector)
	prometheus.MustRegister(buildInfoGauge)

	return collector
}

func (c *Collector) metricsList() []prometheus.GaugeVec {
	return []prometheus.GaugeVec{
		c.WaybackPgs,
		c.PlaybackPgs,
		c.PublishPgs,
	}
}

var startTime time.Time

func init() {
	startTime = time.Now()
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	// Set uptime
	duration := fmt.Sprint(time.Since(startTime).Truncate(time.Second))
	m, err := prometheus.NewConstMetric(c.uptimeDesc, prometheus.GaugeValue, float64(1), duration)
	if err != nil {
		return err
	}
	ch <- m

	return nil
}

// Describe sends the descriptors of each Collector related metrics we have defined
// to the provided prometheus channel.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	logger.Debug("defined wayback metrics")

	for _, metric := range c.metricsList() {
		metric.Describe(ch)
	}
}

// Collect sends all the collected metrics to the provided prometheus channel.
// It requires the caller to handle synchronization.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	logger.Debug("collecting wayback stats")
	if err := c.collect(ch); err != nil {
		logger.Error("error collecting metrics: %v", err)
		return
	}

	for _, metric := range c.metricsList() {
		metric.Collect(ch)
	}
}
