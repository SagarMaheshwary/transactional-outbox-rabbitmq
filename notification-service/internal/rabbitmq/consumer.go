package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/logger"
)

func (r *RabbitMQ) Consume(ctx context.Context, queue string, routingKeys []string) error {
	ch, err := r.NewChannel()
	if err != nil {
		return err
	}

	q, err := r.declareAndBindQueue(ch, queue, routingKeys)
	if err != nil {
		return err
	}

	messages, err := ch.Consume(
		q.Name,
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

	r.Log.Info("AMQP listening for messages", logger.Field{Key: "queue", Value: queue})

	go r.processMessages(ctx, messages)

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(ch *amqp091.Channel, queue string, routingKeys []string) (amqp091.Queue, error) {
	q, err := ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return q, err
	}

	for _, key := range routingKeys {
		if err := ch.QueueBind(q.Name, key, r.Config.Exchange, false, nil); err != nil {
			return q, err
		}
	}

	return q, nil
}

func (r *RabbitMQ) processMessages(ctx context.Context, messages <-chan amqp091.Delivery) {
	for message := range messages {
		r.Log.Info("Broker Message Arrived")

		if err := r.handleMessage(ctx, message); err != nil {
			message.Nack(false, false)
			continue
		}

		message.Ack(false)
	}
}

func (r *RabbitMQ) handleMessage(ctx context.Context, message amqp091.Delivery) error {
	var body map[string]any
	if err := json.Unmarshal(message.Body, &body); err != nil {
		return err
	}

	messageID, ok := message.Headers["message_id"].(string)
	if !ok {
		return errors.New("invalid message id")
	}

	// Try insert into processed_messages to ensure exactly-once processing
	inserted, err := r.ProcessedMessageService.TryInsert(ctx, &model.ProcessedMessage{
		MessageID:   messageID,
		ProcessedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	// If not inserted, message was already processed, skip
	if !inserted {
		return nil
	}

	r.Log.Info("Order email sent to customer", logger.Field{Key: "payload", Value: body})

	return nil
}
