package rabbitmq

import (
	"context"
	"maps"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/observability/metrics"
)

type RetryConfig struct {
	Levels []RetryLevel
}

type RetryLevel struct {
	Name  string
	Delay time.Duration
}

const (
	HeaderRetryCount = "x-retry-count"
)

func (r *RabbitMQ) retryMessage(
	ctx context.Context,
	ch *amqp091.Channel,
	message amqp091.Delivery,
) error {
	retry := r.nextRetryLevel(message.Headers)
	if retry == nil {
		metrics.ConsumerRetryExhaustedTotal.Inc()
		r.Log.WithContext(ctx).Info("Max retry attempts reached, sending message to DLQ",
			logger.Field{Key: "message_id", Value: message.MessageId},
		)
		return r.sendToDLQ(ctx, ch, message)
	}

	metrics.ConsumerRetriesTotal.Inc()

	headers := amqp091.Table{}
	maps.Copy(headers, message.Headers)

	retryCount := getRetryCount(message.Headers)
	headers[HeaderRetryCount] = int32(retryCount + 1)

	r.Log.WithContext(ctx).Info("Sending message to retry exchange",
		logger.Field{Key: "message_id", Value: message.MessageId},
		logger.Field{Key: "retry_level", Value: retry.Name},
		logger.Field{Key: "retry_count", Value: retryCount + 1},
	)

	opts := &PublishOpts{
		Ch:         ch,
		Exchange:   r.Config.Exchange + "." + retry.Name,
		RoutingKey: message.RoutingKey,
		Body:       message.Body,
		Headers:    headers,
		MessageID:  message.MessageId,
	}

	return r.Publish(ctx, opts)
}

func (r *RabbitMQ) sendToDLQ(ctx context.Context, ch *amqp091.Channel, message amqp091.Delivery) error {
	opts := &PublishOpts{
		Ch:         ch,
		Exchange:   r.Config.DLX,
		RoutingKey: r.Config.DLQ,
		Body:       message.Body,
		Headers:    message.Headers,
		MessageID:  message.MessageId,
	}
	if err := r.Publish(ctx, opts); err != nil {
		metrics.ConsumerDLQPublishFailedTotal.Inc()
		r.Log.WithContext(ctx).Error("Failed to publish message to DLQ",
			logger.Field{Key: "message_id", Value: message.MessageId},
			logger.Field{Key: "error", Value: err},
		)
	}

	metrics.ConsumerDLQPublishedTotal.Inc()
	return nil
}

func (r *RabbitMQ) nextRetryLevel(headers amqp091.Table) *RetryLevel {
	count := getRetryCount(headers)
	if count >= len(r.RetryConfig.Levels) {
		return nil
	}
	return &r.RetryConfig.Levels[count]
}
