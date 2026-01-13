package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
)

func (r *RabbitMQ) Health() error {
	if r.Conn == nil || r.Conn.IsClosed() {
		return errors.New("RabbitMQ healthcheck failed")
	}

	return nil
}

func (r *RabbitMQ) Close() error {
	if r.Conn != nil && !r.Conn.IsClosed() {
		return r.Conn.Close()
	}
	return nil
}

func (r *RabbitMQ) connect(ctx context.Context) error {
	var err error

	address := fmt.Sprintf("amqp://%s:%s@%s:%d", r.Config.Username, r.Config.Password, r.Config.Host, r.Config.Port)
	r.Conn, err = amqp091.Dial(address)
	if err != nil {
		r.Log.Error("RabbitMQ connection error", logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	r.Log.Info("RabbitMQ connected")

	channel, err := r.NewChannel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := r.initExchanges(channel); err != nil {
		return err
	}
	if err := r.initDLQQueue(channel); err != nil {
		return err
	}

	return nil
}
