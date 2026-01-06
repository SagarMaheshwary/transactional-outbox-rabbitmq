package rabbitmq

import (
	"context"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/config"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
)

type RabbitMQService interface {
	Health() error
	Close() error
	NewChannel() (*amqp091.Channel, error)
	Publish(ctx context.Context, ch *amqp091.Channel, routingKey string, message interface{}, messageID string) error
	initExchange(ch *amqp091.Channel) error
	connect(ctx context.Context) error
}

type RabbitMQ struct {
	Config        *config.AMQP
	Conn          *amqp091.Connection
	reconnectLock sync.Mutex
	Log           logger.Logger
}

type Opts struct {
	Config *config.AMQP
	Logger logger.Logger
}

func NewRabbitMQ(ctx context.Context, opts *Opts) (RabbitMQService, error) {
	b := &RabbitMQ{
		Config: opts.Config,
		Log:    opts.Logger,
	}
	if err := b.connect(ctx); err != nil {
		return nil, err
	}

	return b, nil
}

func (r *RabbitMQ) NewChannel() (*amqp091.Channel, error) {
	c, err := r.Conn.Channel()

	if err != nil {
		r.Log.Error("RabbitMQ channel error", logger.Field{Key: "error", Value: err.Error()})
		return nil, err
	}

	return c, nil
}
