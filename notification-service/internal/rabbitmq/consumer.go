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
		r.Log.Info("Broker Message Arrived", logger.Field{Key: "routing_key", Value: message.RoutingKey}, logger.Field{Key: "headers", Value: message.Headers})

		err := r.handleMessage(ctx, message)
		if err == nil {
			message.Ack(false)
			continue
		}

		if errors.Is(err, constant.ErrPermanent) {
			r.Log.Error("Permanent error processing message, sending to DLQ", logger.Field{Key: "error", Value: err.Error()})
			r.sendToDLQ(ctx, ch, message)
		}
		if errors.Is(err, constant.ErrTransient) {
			r.Log.Error("Transient error processing message, sending to retry exchange", logger.Field{Key: "error", Value: err.Error()})
			r.RetryMessage(ctx, ch, message)
		}

		message.Ack(false)
	}
}

func (r *RabbitMQ) handleMessage(ctx context.Context, message amqp091.Delivery) error {
	ctx = contextWithOtelHeaders(ctx, message.Headers)
	ctx, span := tracing.Tracer.Start(ctx, "rabbitmq.consume")
	defer span.End()

	var body map[string]any
	if err := json.Unmarshal(message.Body, &body); err != nil {
		return fmt.Errorf("%w: failed to unmarshal message body: %v", constant.ErrPermanent, err)
	}

	span.SetAttributes(attribute.KeyValue{Key: "rabbitmq.routing_key", Value: attribute.StringValue(message.RoutingKey)})

	messageID, ok := message.Headers["message_id"].(string)
	if !ok {
		return fmt.Errorf("%w: message_id header missing or invalid", constant.ErrPermanent)
	}

	// As it's a demo, we are inlining the processing logic here. In real-world applications,
	// consider delegating to a dedicated service method.
	err := service.WithTransaction(ctx, r.DB, func(tx *gorm.DB) error {
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

		// Idempotency: message already processed → commit & ack
		if !inserted {
			return nil
		}

		// Here we are just logging the email sending action. In real-world applications,
		// you would have business logic with database queries that would
		// be part of the same transaction as the processed message.
		r.Log.Info(
			"Order email sent to customer",
			logger.Field{Key: "payload", Value: body},
		)

		return nil
	})
	if err != nil {
		return fmt.Errorf("%w: failed to process message: %v", constant.ErrTransient, err)
	}

	return nil
}
