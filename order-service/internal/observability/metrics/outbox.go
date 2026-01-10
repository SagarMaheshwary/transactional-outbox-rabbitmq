package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	OutboxBacklog = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_backlog",
		Help: "Number of outbox events waiting to be processed.",
	})
	OutboxEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "outbox_events_total",
			Help: "Total number of outbox events processed.",
		},
		[]string{"status"}, // published | failed
	)
	OutboxPublishLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "outbox_publish_latency_seconds",
			Help:    "End-to-end latency from outbox insert to successful publish.",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
	)
)

type OutboxEventMetrics struct{}

func (OutboxEventMetrics) Register(r *prometheus.Registry) {
	r.MustRegister(
		OutboxBacklog,
		OutboxEventsTotal,
		OutboxPublishLatency,
	)
}
