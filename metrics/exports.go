// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package metrics // import "github.com/wabarc/wayback/metrics"

import (
	"bytes"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/wabarc/logger"
)

// Gather holds configured collector.
var Gather *Collector

// Export export metrics by the given labels for matching.
func (c *Collector) Export(labels ...string) string {
	logger.Debug("export metrics family: %#v", prometheus.DefaultRegisterer)

	var gatherer = prometheus.DefaultGatherer
	var protobufs, err = gatherer.Gather()
	if err != nil {
		logger.Error("gather metrics family failed: %v", err)
	}

	var s string
	for _, pb := range protobufs {
		var buf bytes.Buffer
		if _, err := expfmt.MetricFamilyToText(&buf, pb); err != nil {
			logger.Error("export to text failed: %v", err)
		}
		logger.Debug("string: %v\nname: %v\nhelp: %v\ntype: %v\nmetric: %v\nvalue: %v",
			buf.String(), pb.GetName(), pb.GetHelp(), pb.GetType(), pb.GetMetric(), pb.GetMetric()[0].GetGauge().GetValue())
		if match(pb.GetName(), labels...) {
			s += buf.String()
		}
	}

	return s
}

func match(name string, labels ...string) bool {
	for _, label := range labels {
		if strings.HasPrefix(name, label) {
			return true
		}
	}
	return false
}
