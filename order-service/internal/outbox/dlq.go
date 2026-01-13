package outbox

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/rabbitmq"
)

func (o *Outbox) publishToDLQ(
	ctx context.Context,
	ch *amqp091.Channel,
	event *model.OutboxEvent,
	procErr error,
) {
	dlqEvent := map[string]interface{}{
		"event_id":       event.ID,
		"event_key":      event.EventKey,
		"payload":        event.Payload,
		"failed_at":      time.Now(),
		"failure_reason": procErr.Error(),
	}

	err := o.rabbitmq.Publish(
		ctx,
		&rabbitmq.PublishOpts{
			Ch:         ch,
			Exchange:   o.amqpConfig.DLX,
			RoutingKey: o.amqpConfig.DLQ,
			Body:       dlqEvent,
			Headers:    amqp091.Table{"message_id": event.ID},
		},
	)
	if err != nil {
		o.log.Error("Failed to publish event to DLQ", logger.Field{Key: "error", Value: err.Error()})
		return
	}

	o.log.Info("Event sent to DLQ after max retries",
		logger.Field{Key: "event_id", Value: event.ID},
		logger.Field{Key: "event_key", Value: event.EventKey},
		logger.Field{Key: "error", Value: procErr.Error()},
	)
}

func (o *Outbox) markFailed(
	ctx context.Context,
	event *model.OutboxEvent,
	procErr error,
) {
	err := o.outboxEventService.UpdateState(ctx, event.ID, map[string]interface{}{
		"status":         model.OutboxEventStatusFailed,
		"failure_reason": procErr.Error(),
		"failed_at":      time.Now(),
		"locked_at":      nil,
		"locked_by":      nil,
	})
	if err != nil {
		o.log.Error("Failed to update event status to Failed",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "event_id", Value: event.ID},
		)
	}
}
