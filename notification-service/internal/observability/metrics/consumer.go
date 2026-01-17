package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ConsumerMessagesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "consumer_messages_total",
		Help: "Total number of messages received by the consumer.",
	})
	ConsumerProcessingFailedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "consumer_processing_failed_total",
		Help: "Total number of messages that failed during processing.",
	})
	ConsumerRetriesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "consumer_retries_total",
		Help: "Total number of consumer message processing retries.",
	})
	ConsumerRetryExhaustedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "consumer_retry_exhausted_total",
		Help: "Total number of consumer messages that have exhausted all retry attempts.",
	})
	ConsumerDLQPublishedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "consumer_dlq_published_total",
		Help: "Total number of consumer messages published to the dead-letter queue.",
	})
	ConsumerDLQPublishFailedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "consumer_dlq_publish_failed_total",
		Help: "Total number of consumer messages that failed to publish to the dead-letter queue.",
	})
)

type ConsumerMetrics struct{}

func (ConsumerMetrics) Register(r *prometheus.Registry) {
	r.MustRegister(
		ConsumerMessagesTotal,
		ConsumerProcessingFailedTotal,
		ConsumerRetriesTotal,
		ConsumerRetryExhaustedTotal,
		ConsumerDLQPublishedTotal,
		ConsumerDLQPublishFailedTotal,
	)
}
