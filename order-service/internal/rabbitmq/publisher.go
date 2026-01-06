package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
)

func (b *RabbitMQ) Publish(
	ctx context.Context,
	ch *amqp091.Channel,
	routingKey string,
	message interface{},
	messageID string,
) error {
	ctx, cancel := context.WithTimeout(ctx, b.Config.PublishTimeout)
	defer cancel()

	messageData, err := json.Marshal(&message)
	if err != nil {
		return err
	}

	headers := amqp091.Table{}
	headers["message_id"] = messageID

	err = ch.PublishWithContext(
		ctx,
		b.Config.Exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        messageData,
			Headers:     headers,
		},
	)
	if err != nil {
		b.Log.Error("RabbitMQ failed to publish message",
			logger.Field{Key: "routing_key", Value: routingKey},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return err
	}

	b.Log.Info("RabbitMQ message published", logger.Field{Key: "routing_key", Value: routingKey})

	return nil
}
