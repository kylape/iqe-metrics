// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// A minimal example of how to include Prometheus instrumentation.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

var labels = []string{"plugin", "priority", "target_app"}

var ran = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "iqe",
		Subsystem: "tests",
		Name:      "rans",
		Help:      "Running count of tests that have ran",
	},
	labels,
)

var failed = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "iqe",
		Subsystem: "tests",
		Name:      "failed",
		Help:      "Running count of tests that have failed",
	},
	labels,
)

var skipped = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "iqe",
		Subsystem: "tests",
		Name:      "skipped",
		Help:      "Running count of tests that have been skipped",
	},
	labels,
)

var errored = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "iqe",
		Subsystem: "tests",
		Name:      "errors",
		Help:      "Running count of tests that have errored",
	},
	labels,
)

var time = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "iqe",
		Subsystem: "tests",
		Name:      "time_seconds",
		Help:      "Running count of seconds that have executed",
	},
	labels,
)

type MetricsResult struct {
	Plugin    string `json:"plugin"`
	Priority  string `json:"priority"`
	TargetApp string `json:"targetApp"`
	Results   struct {
		Ran     int `json:"ran"`
		Failed  int `json:"failed"`
		Skipped int `json:"skipped"`
		Errored int `json:"errored"`
	} `json:"results"`
	Time float64 `json:"time"`
}

func main() {
	flag.Parse()
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/results", handleResults)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func handleResults(w http.ResponseWriter, r *http.Request) {
	var mr MetricsResult

	err := json.NewDecoder(r.Body).Decode(&mr)

	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
	}

	ran.WithLabelValues(mr.Plugin, mr.Priority, mr.TargetApp).Add(float64(mr.Results.Ran))
	failed.WithLabelValues(mr.Plugin, mr.Priority, mr.TargetApp).Add(float64(mr.Results.Failed))
	skipped.WithLabelValues(mr.Plugin, mr.Priority, mr.TargetApp).Add(float64(mr.Results.Skipped))
	errored.WithLabelValues(mr.Plugin, mr.Priority, mr.TargetApp).Add(float64(mr.Results.Errored))
	time.WithLabelValues(mr.Plugin, mr.Priority, mr.TargetApp).Add(mr.Time)
	w.WriteHeader(200)
}

func init() {
	// Register the summary and the histogram with Prometheus's default registry.
	prometheus.MustRegister(ran)
	prometheus.MustRegister(failed)
	prometheus.MustRegister(skipped)
	prometheus.MustRegister(errored)
	prometheus.MustRegister(time)
	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}
