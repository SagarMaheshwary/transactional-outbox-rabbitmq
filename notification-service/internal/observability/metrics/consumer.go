package metrics

import "github.com/prometheus/client_golang/prometheus"

type ConsumerMetrics struct{}

func (ConsumerMetrics) Register(r *prometheus.Registry) {
	// r.MustRegister(
	// )
}
