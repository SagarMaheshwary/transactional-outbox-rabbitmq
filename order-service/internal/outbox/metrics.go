package outbox

import (
	"context"
	"time"

	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/observability/metrics"
)

func (o *Outbox) startMetricsReporter(ctx context.Context) {
	ticker := time.NewTicker(o.config.BacklogReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			o.reportBacklog(ctx)
			o.reportRetryBacklog(ctx)
		}
	}
}

func (o *Outbox) reportBacklog(ctx context.Context) {
	count, err := o.outboxEventService.CountBacklog(ctx)
	if err != nil {
		o.log.Error("Failed to count outbox backlog",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return
	}
	metrics.OutboxBacklog.Set(float64(count))
}

func (o *Outbox) reportRetryBacklog(ctx context.Context) {
	count, err := o.outboxEventService.CountWaitingRetry(ctx)
	if err != nil {
		o.log.Warn("Failed to count retry backlog",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return
	}
	metrics.OutboxEventsWaitingRetry.Set(float64(count))
}
