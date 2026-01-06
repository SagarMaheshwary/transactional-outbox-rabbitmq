package rabbitmq

import "github.com/rabbitmq/amqp091-go"

func (b *RabbitMQ) initExchange(ch *amqp091.Channel) error {
	return ch.ExchangeDeclare(
		b.Config.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}
