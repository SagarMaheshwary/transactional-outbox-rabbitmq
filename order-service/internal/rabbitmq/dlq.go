package rabbitmq

import "github.com/rabbitmq/amqp091-go"

func (r *RabbitMQ) initDLQQueue(ch *amqp091.Channel) error {
	_, err := ch.QueueDeclare(
		r.Config.DLQ,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		r.Config.DLQ,
		r.Config.DLQ,
		r.Config.DLX,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
