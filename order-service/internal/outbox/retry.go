package outbox

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
)

func (o *Outbox) handleFailure(
	ctx context.Context,
	ch *amqp091.Channel,
	event *model.OutboxEvent,
	err error,
) {
	event.RetryCount++

	if event.RetryCount > o.config.MaxRetryCount {
		o.markFailed(ctx, event, err)
		o.publishToDLQ(ctx, ch, event, err)
		return
	}

	o.scheduleRetry(ctx, event)
}

func (o *Outbox) scheduleRetry(ctx context.Context, event *model.OutboxEvent) {
	backoff := backoff(event.RetryCount, o.config.RetryDelay)
	o.log.Info("Scheduling retry for event",
		logger.Field{Key: "event_id", Value: event.ID},
		logger.Field{Key: "retry_count", Value: event.RetryCount},
		logger.Field{Key: "backoff_seconds", Value: backoff.Seconds()},
	)

	err := o.outboxEventService.UpdateState(ctx, event.ID, map[string]interface{}{
		"status":        model.OutboxEventStatusPending,
		"retry_count":   event.RetryCount,
		"next_retry_at": time.Now().Add(backoff),
		"locked_at":     nil,
		"locked_by":     nil,
	})
	if err != nil {
		o.log.Error("Failed to schedule retry for event", logger.Field{Key: "error", Value: err.Error()})
		return
	}
}
