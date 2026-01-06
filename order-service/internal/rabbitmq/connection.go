package rabbitmq

import (
	"errors"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
)

func (b *RabbitMQ) Health() error {
	if b.Conn == nil || b.Conn.IsClosed() {
		return errors.New("RabbitMQ healthcheck failed")
	}

	return nil
}

func (b *RabbitMQ) Close() error {
	if b.Conn != nil && !b.Conn.IsClosed() {
		return b.Conn.Close()
	}
	return nil
}

func (b *RabbitMQ) connect() error {
	var err error

	address := fmt.Sprintf("amqp://%s:%s@%s:%d", b.Config.Username, b.Config.Password, b.Config.Host, b.Config.Port)
	b.Conn, err = amqp091.Dial(address)
	if err != nil {
		b.Log.Error("RabbitMQ connection error", logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	b.Log.Info("RabbitMQ connected")

	channel, err := b.NewChannel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := b.initExchange(channel); err != nil {
		return err
	}

	return nil
}
