package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/logger"
)

type PublishOpts struct {
	Ch         *amqp091.Channel
	Exchange   string
	RoutingKey string
	Body       interface{}
	Headers    amqp091.Table
	MessageID  string
}

func (r *RabbitMQ) Publish(ctx context.Context, opts *PublishOpts) error {
	ctx, cancel := context.WithTimeout(ctx, r.Config.PublishTimeout)
	defer cancel()

	var body []byte
	if b, ok := opts.Body.([]byte); ok {
		body = b
	} else {
		var err error
		body, err = json.Marshal(opts.Body)
		if err != nil {
			return err
		}
	}

	err := opts.Ch.PublishWithContext(
		ctx,
		opts.Exchange,
		opts.RoutingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers:     opts.Headers,
			MessageId:   opts.MessageID,
		},
	)
	if err != nil {
		r.Log.Error("RabbitMQ failed to publish message",
			logger.Field{Key: "routing_key", Value: opts.RoutingKey},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return err
	}

	r.Log.Info("RabbitMQ message published", logger.Field{Key: "routing_key", Value: opts.RoutingKey})

	return nil
}
