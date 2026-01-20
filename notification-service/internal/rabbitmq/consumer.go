package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/constant"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/observability/metrics"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/observability/tracing"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/service"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

func (r *RabbitMQ) Consume(ctx context.Context) error {
	ch, err := r.NewChannel()
	if err != nil {
		return err
	}

	messages, err := ch.Consume(
		r.Config.Queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	r.Log.Info("AMQP listening for messages", logger.Field{Key: "queue", Value: r.Config.Queue})

	go r.processMessages(ctx, ch, messages)

	return nil
}

func (r *RabbitMQ) processMessages(ctx context.Context, ch *amqp091.Channel, messages <-chan amqp091.Delivery) {
	for message := range messages {
		metrics.ConsumerMessagesTotal.Inc()

		retryCount := getRetryCount(message.Headers)
		r.Log.WithContext(ctx).Info("Broker Message Arrived",
			logger.Field{Key: "routing_key", Value: message.RoutingKey},
			logger.Field{Key: "message_id", Value: message.MessageId},
			logger.Field{Key: "retry_count", Value: retryCount},
		)

		ctx = contextWithOtelHeaders(ctx, message.Headers)
		ctx, span := tracing.Tracer.Start(ctx, "rabbitmq.consume")
		span.SetAttributes(
			attribute.String("rabbitmq.routing_key", message.RoutingKey),
			attribute.Int("rabbitmq.retry_count", retryCount),
		)

		outcome := "success"

		err := r.handleMessage(ctx, message)
		if err != nil {
			metrics.ConsumerProcessingFailedTotal.Inc()

			if errors.Is(err, constant.ErrPermanent) {
				r.Log.WithContext(ctx).Error("Permanent error processing message, sending to DLQ",
					logger.Field{Key: "error", Value: err.Error()},
					logger.Field{Key: "message_id", Value: message.MessageId},
				)
				r.sendToDLQ(ctx, ch, message)
				outcome = "dlq"
			} else if errors.Is(err, constant.ErrTransient) {
				r.Log.WithContext(ctx).Error("Transient error processing message, sending to retry exchange",
					logger.Field{Key: "error", Value: err.Error()},
					logger.Field{Key: "message_id", Value: message.MessageId},
				)
				r.retryMessage(ctx, ch, message)
				outcome = "retry"
			} else {
				span.RecordError(err)
				outcome = "failed"
			}
		}

		span.SetAttributes(attribute.String("consumer.outcome", outcome))
		span.End()

		message.Ack(false)
	}
}

func (r *RabbitMQ) handleMessage(ctx context.Context, message amqp091.Delivery) error {
	var body map[string]any
	if err := json.Unmarshal(message.Body, &body); err != nil {
		return fmt.Errorf("%w: failed to unmarshal message body: %v",
			constant.ErrPermanent, err,
		)
	}

	if message.MessageId == "" {
		err := fmt.Errorf("%w: message_id header missing or invalid",
			constant.ErrPermanent,
		)
		return err
	}

	err := service.WithTransaction(ctx, r.DB, func(tx *gorm.DB) error {
		return r.processMessageOnce(ctx, tx, message.MessageId, body)
	})
	if err != nil {
		return fmt.Errorf("%w: failed to process message: %v",
			constant.ErrTransient, err,
		)
	}

	return nil
}

func (r *RabbitMQ) processMessageOnce(
	ctx context.Context,
	tx *gorm.DB,
	messageID string,
	body map[string]any,
) error {
	inserted, err := r.ProcessedMessageService.TryInsert(
		ctx,
		tx,
		&model.ProcessedMessage{
			MessageID:   messageID,
			ProcessedAt: time.Now(),
		},
	)
	if err != nil {
		return err
	}

	// Idempotency: already processed → no-op
	if !inserted {
		return nil
	}

	// Business logic (demo)
	r.Log.Info(
		"Order email sent to customer",
		logger.Field{Key: "payload", Value: body},
	)

	return nil
}
