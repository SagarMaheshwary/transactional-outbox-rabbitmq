package outbox

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
)

func (o *Outbox) processEvents(
	ctx context.Context,
	workerID int,
	ch *amqp091.Channel,
	events <-chan *model.OutboxEvent,
) {
	for event := range events {
		o.log.Info("Worker processing event",
			logger.Field{Key: "worker_id", Value: workerID},
			logger.Field{Key: "event_id", Value: event.ID},
			logger.Field{Key: "event_key", Value: event.EventKey},
		)

		err := o.PublishEvent(ctx, ch, event)
		if err == nil {
			o.markPublished(ctx, event)
			continue
		}

		o.markFailed(ctx, event, err)
	}
}

func (o *Outbox) PublishEvent(ctx context.Context, ch *amqp091.Channel, event *model.OutboxEvent) error {
	err := o.rabbitmq.Publish(ctx, ch, event.EventKey, event.Payload, event.ID)
	if err != nil {
		o.log.Error("Failed to publish event", logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	return nil
}

func (o *Outbox) markPublished(ctx context.Context, event *model.OutboxEvent) {
	err := o.outboxEventService.UpdateState(ctx, event.ID, map[string]interface{}{
		"status":    model.OutboxEventStatusPublished,
		"locked_at": nil,
		"locked_by": nil,
	})
	if err != nil {
		o.log.Error("Failed to update event status to Published", logger.Field{Key: "error", Value: err.Error()})
		return
	}
}

func (o *Outbox) markFailed(ctx context.Context, event *model.OutboxEvent, procErr error) {
	err := o.outboxEventService.UpdateState(ctx, event.ID, map[string]interface{}{
		"status":         model.OutboxEventStatusFailed,
		"failure_reason": procErr.Error(),
		"failed_at":      time.Now(),
		"locked_at":      nil,
		"locked_by":      nil,
	})
	if err != nil {
		o.log.Error("Failed to update event status to Failed", logger.Field{Key: "error", Value: err.Error()})
		return
	}
}
