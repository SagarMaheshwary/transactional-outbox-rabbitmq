package outbox

import (
	"context"
	"time"

	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/observability/metrics"
)

func (o *Outbox) startBacklogReporter(ctx context.Context) {
	ticker := time.NewTicker(o.config.BacklogReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := o.outboxEventService.CountBacklog(ctx)
			if err != nil {
				o.log.Error("Failed to count outbox backlog",
					logger.Field{Key: "error", Value: err.Error()},
				)
				continue
			}
			metrics.OutboxBacklog.Set(float64(count))
		}
	}
}
