package outbox

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/observability/metrics"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/observability/tracing"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/rabbitmq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

		if event.Traceparent != "" {
			ctx = tracing.ExtractTraceParent(ctx, event.Traceparent)
		}

		ctx, span := tracing.Tracer.Start(
			ctx,
			"Outbox.PublishEvent",
			trace.WithAttributes(
				attribute.String("event.id", event.ID),
				attribute.String("event.key", event.EventKey),
				attribute.Int("event.retry_count", event.RetryCount),
			),
		)

		err := o.PublishEvent(ctx, ch, event)

		outcome := "published"
		if err != nil {
			outcome = o.handleFailure(ctx, ch, event, err)
		}

		o.markPublished(ctx, event)

		span.SetAttributes(attribute.String("outbox.outcome", outcome))
		span.End()
	}
}

func (o *Outbox) PublishEvent(
	ctx context.Context,
	ch *amqp091.Channel,
	event *model.OutboxEvent,
) error {
	o.log.WithContext(ctx).Info("Publishing outbox event",
		logger.Field{Key: "event_id", Value: event.ID},
		logger.Field{Key: "event_key", Value: event.EventKey},
		logger.Field{Key: "retry_count", Value: event.RetryCount},
	)

	err := o.rabbitmq.Publish(
		ctx,
		&rabbitmq.PublishOpts{
			Ch:         ch,
			Exchange:   o.amqpConfig.Exchange,
			RoutingKey: event.EventKey,
			Body:       event.Payload,
			MessageID:  event.ID,
		},
	)
	if err != nil {
		metrics.OutboxEventsTotal.WithLabelValues("failed").Inc()
		o.log.WithContext(ctx).Error("Failed to publish outbox event",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "event_id", Value: event.ID},
		)
		return err
	}

	latency := time.Since(event.CreatedAt).Seconds()
	metrics.OutboxPublishLatency.Observe(latency)
	metrics.OutboxEventsTotal.WithLabelValues("published").Inc()

	return nil
}

func (o *Outbox) markPublished(ctx context.Context, event *model.OutboxEvent) {
	_, err := o.outboxEventService.UpdateStateIfInProgress(ctx, event.ID, map[string]interface{}{
		"status":    model.OutboxEventStatusPublished,
		"locked_at": nil,
		"locked_by": nil,
	})
	if err != nil {
		o.log.WithContext(ctx).Error("Failed to update event status to Published",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "event_id", Value: event.ID},
		)
	}
}
