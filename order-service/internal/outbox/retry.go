package outbox

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/observability/metrics"
)

func (o *Outbox) handleFailure(
	ctx context.Context,
	ch *amqp091.Channel,
	event *model.OutboxEvent,
	err error,
) string {
	if event.RetryCount >= o.config.MaxRetryCount {
		metrics.OutboxRetryExhaustionsTotal.Inc()
		if markErr := o.markFailed(ctx, event, err); markErr != nil {
			return "failed"
		}
		o.publishToDLQ(ctx, ch, event, err)
		return "dlq"
	}

	metrics.OutboxRetriesTotal.Inc()
	event.RetryCount++
	if scheduleErr := o.scheduleRetry(ctx, event); scheduleErr != nil {
		return "failed"
	}
	return "retry"
}

func (o *Outbox) scheduleRetry(ctx context.Context, event *model.OutboxEvent) error {
	backoff := backoff(event.RetryCount, o.config.RetryDelay)
	o.log.WithContext(ctx).Info("Scheduling retry for event",
		logger.Field{Key: "event_id", Value: event.ID},
		logger.Field{Key: "retry_count", Value: event.RetryCount},
		logger.Field{Key: "backoff_seconds", Value: backoff.Seconds()},
	)

	_, err := o.outboxEventService.UpdateStateIfInProgress(ctx, event.ID, map[string]interface{}{
		"status":        model.OutboxEventStatusPending,
		"retry_count":   event.RetryCount,
		"next_retry_at": time.Now().Add(backoff),
		"locked_at":     nil,
		"locked_by":     nil,
	})
	if err != nil {
		o.log.WithContext(ctx).Error("Failed to schedule retry for event",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "event_id", Value: event.ID},
		)
		return err
	}

	return nil
}
