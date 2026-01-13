package rabbitmq

import "github.com/rabbitmq/amqp091-go"

func (r *RabbitMQ) initExchanges(ch *amqp091.Channel) error {
	exchanges := []struct {
		Name string
		Kind string
	}{
		{Name: r.Config.Exchange, Kind: "topic"},
		{Name: r.Config.DLX, Kind: "direct"},
	}

	for _, exchange := range exchanges {
		err := ch.ExchangeDeclare(
			exchange.Name,
			exchange.Kind,
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
