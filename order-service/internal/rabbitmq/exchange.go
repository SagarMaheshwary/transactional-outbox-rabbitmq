package rabbitmq

import "github.com/rabbitmq/amqp091-go"

func (r *RabbitMQ) initExchange(ch *amqp091.Channel) error {
	exchanges := []string{
		r.Config.Exchange,
		r.Config.DLX,
	}

	for _, exchange := range exchanges {
		err := ch.ExchangeDeclare(
			exchange,
			"topic",
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

	return nil
}
