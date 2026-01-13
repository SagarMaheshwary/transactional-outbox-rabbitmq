package rabbitmq

import (
	"github.com/rabbitmq/amqp091-go"
)

func (r *RabbitMQ) initQueue(ch *amqp091.Channel) error {
	q, err := ch.QueueDeclare(
		r.Config.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for _, key := range RoutingKeys {
		if err := ch.QueueBind(q.Name, key, r.Config.Exchange, false, nil); err != nil {
			return err
		}
	}

	return nil
}

func (r *RabbitMQ) initRetryExchangeAndQueues(ch *amqp091.Channel) error {
	for _, retry := range r.RetryConfig.Levels {
		err := ch.ExchangeDeclare(
			r.Config.Exchange+"."+retry.Name,
			"direct",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}

	for _, retry := range r.RetryConfig.Levels {
		_, err := ch.QueueDeclare(
			r.Config.Queue+"."+retry.Name,
			true,
			false,
			false,
			false,
			amqp091.Table{
				"x-dead-letter-exchange": r.Config.Exchange,
				"x-message-ttl":          retry.Delay.Milliseconds(),
			},
		)
		if err != nil {
			return err
		}

		for _, routingKey := range RoutingKeys {
			err = ch.QueueBind(
				r.Config.Queue+"."+retry.Name,
				routingKey,
				r.Config.Exchange+"."+retry.Name,
				false,
				nil,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
