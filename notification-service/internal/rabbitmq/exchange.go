package rabbitmq

import "github.com/rabbitmq/amqp091-go"

func (r *RabbitMQ) initExchange(ch *amqp091.Channel) error {
	return ch.ExchangeDeclare(
		r.Config.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}
