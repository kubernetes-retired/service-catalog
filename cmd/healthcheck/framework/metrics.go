/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var registerMetrics sync.Once

const (
	promNamespace = "servicecatalog_health" // Prometheus namespace (nothing to do with k8s namespace)
)

var (
	// Metrics are identified in Prometheus by concatinating Namespace,
	// Subsystem and Name while omitting any nulls and separating each key with
	// an underscore.  Note that in this context, Namespace is the Prometheus
	// Namespace and there is no correlation with Kubernetes Namespace.

	// ExecutionCount is the number of times the HealthCheck has executed
	ExecutionCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: promNamespace,
			Name:      "execution_count",
			Help:      "Number of times the health check has run.",
		},
	)

	// ErrorCount is the number of times HealthCheck has errored during the end to end test
	ErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: promNamespace,
			Name:      "error_count",
			Help:      "Number of times the health check ended in error, by error.",
		},
		[]string{"error"},
	)

	// eventHandlingTime is a histogram recording how long a operation took
	eventHandlingTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  promNamespace,
			Name:       "successful_duration_seconds",
			Help:       "processing time (s) of successfully executed operation, by operation.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		}, []string{"operation"})
)

// ReportOperationCompleted records the elapses time in milliseconds for a specified operation
func ReportOperationCompleted(operation string, startTime time.Time) {
	eventHandlingTimeSummary.WithLabelValues(operation).Observe(time.Since(startTime).Seconds())
}

func register(registry *prometheus.Registry) {
	registerMetrics.Do(func() {
		registry.MustRegister(ExecutionCount)
		registry.MustRegister(ErrorCount)
		registry.MustRegister(eventHandlingTimeSummary)
	})
}

// RegisterMetricsAndInstallHandler registers the metrics objects with
// Prometheus and installs the Prometheus http handler at the default context.
func RegisterMetricsAndInstallHandler(m *http.ServeMux) {
	registry := prometheus.NewRegistry()
	register(registry)
	m.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	glog.V(3).Info("Registered /metrics with prometheus")
}
