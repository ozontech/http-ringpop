package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	NS          = "ringpop"
	MetricsPath = "/metrics"
)

var (
	registerer = prometheus.DefaultRegisterer
)

// Handler returns HTTP handler for serving metrics
func Handler() http.Handler {
	return promhttp.Handler()
}

// MustRegister registers a new metric in the registry. If the metric fail to register it will panic.
func MustRegister(collectors ...prometheus.Collector) {
	registerer.MustRegister(collectors...)
}

// NewCounter creates a new Counter with predefined namespace
func NewCounter(name, help string) prometheus.Counter {
	return prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: NS,
			Name:      name,
			Help:      help,
		},
	)
}

// MustRegisterCounter creates and registers new Counter with predefined namespace
// Panics if metrics with same name already registered
func MustRegisterCounter(name, help string) prometheus.Counter {
	collector := NewCounter(name, help)
	MustRegister(collector)

	return collector
}
