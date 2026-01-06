package outbox

import (
	"context"

	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
)

func (o *Outbox) dispatchPendingEvents(
	ctx context.Context,
	workerID string,
	eventsCh chan<- *model.OutboxEvent,
) {
	events, err := o.outboxEventService.ClaimEvents(ctx, workerID, o.config.BatchSize)
	if err != nil {
		o.log.Info("DB query failed", logger.Field{Key: "error", Value: err.Error()})
		return
	}

	if len(events) == 0 {
		return
	}

	o.log.Info("Fetched outbox events", logger.Field{Key: "count", Value: len(events)})

	for _, event := range events {
		select {
		case <-ctx.Done():
			return
		case eventsCh <- event:
		}
	}
}
