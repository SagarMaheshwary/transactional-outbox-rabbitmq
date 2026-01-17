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
	OutboxRetriesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "outbox_retries_total",
		Help: "Total number of outbox event publish retries.",
	})
	OutboxEventsWaitingRetry = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_events_waiting_retry",
		Help: "Number of outbox events currently waiting to be retried.",
	})
	OutboxRetryExhaustionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "outbox_retry_exhaustions_total",
		Help: "Total number of outbox events that have exhausted all retry attempts.",
	})
	OutboxDLQPublishedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "outbox_dlq_published_total",
		Help: "Total number of outbox events published to the dead-letter queue.",
	})
	OutboxDLQPublishFailedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "outbox_dlq_publish_failed_total",
		Help: "Total number of outbox events that failed to publish to the dead-letter queue.",
	})
)

type OutboxEventMetrics struct{}

func (OutboxEventMetrics) Register(r *prometheus.Registry) {
	r.MustRegister(
		OutboxBacklog,
		OutboxEventsTotal,
		OutboxPublishLatency,
		OutboxRetriesTotal,
		OutboxEventsWaitingRetry,
		OutboxRetryExhaustionsTotal,
		OutboxDLQPublishedTotal,
		OutboxDLQPublishFailedTotal,
	)
}
