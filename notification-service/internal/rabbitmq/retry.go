package rabbitmq

import (
	"context"
	"maps"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/logger"
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

func (r *RabbitMQ) RetryMessage(
	ctx context.Context,
	ch *amqp091.Channel,
	message amqp091.Delivery,
) error {
	retry := r.nextRetryLevel(message.Headers)
	if retry == nil {
		return r.sendToDLQ(ctx, ch, message)
	}

	headers := amqp091.Table{}
	maps.Copy(headers, message.Headers)

	retryCount := getRetryCount(message.Headers)
	r.Log.Info("Current Retry Count",
		logger.Field{Key: "retry_count", Value: retryCount},
		logger.Field{Key: "next_exchange", Value: retry.Name},
	)
	headers[HeaderRetryCount] = int32(retryCount + 1)

	opts := &PublishOpts{
		Ch:         ch,
		Exchange:   r.Config.Exchange + "." + retry.Name,
		RoutingKey: message.RoutingKey,
		Body:       message.Body,
		Headers:    headers,
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
	}

	return r.Publish(ctx, opts)
}

func (r *RabbitMQ) nextRetryLevel(headers amqp091.Table) *RetryLevel {
	count := getRetryCount(headers)
	if count >= len(r.RetryConfig.Levels) {
		return nil
	}
	return &r.RetryConfig.Levels[count]
}
