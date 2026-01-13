package rabbitmq

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/config"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/database"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/service"
	"gorm.io/gorm"
)

type RabbitMQService interface {
	Health() error
	Close() error
	NewChannel() (*amqp091.Channel, error)
	Publish(ctx context.Context, opts *PublishOpts) error
	Consume(ctx context.Context) error
}

type RabbitMQ struct {
	Config                  *config.AMQP
	Conn                    *amqp091.Connection
	Log                     logger.Logger
	ProcessedMessageService service.ProcessedMessageService
	DB                      *gorm.DB
	RetryConfig             RetryConfig
}

type Opts struct {
	Config                  *config.AMQP
	Logger                  logger.Logger
	ProcessedMessageService service.ProcessedMessageService
	DB                      database.DatabaseService
}

func NewRabbitMQ(ctx context.Context, opts *Opts) (RabbitMQService, error) {
	b := &RabbitMQ{
		Config:                  opts.Config,
		Log:                     opts.Logger,
		ProcessedMessageService: opts.ProcessedMessageService,
		DB:                      opts.DB.DB(),
		RetryConfig: RetryConfig{
			Levels: []RetryLevel{
				{"retry.30s", 30 * time.Second},
				{"retry.1m", 1 * time.Minute},
				{"retry.5m", 5 * time.Minute},
			},
		},
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
