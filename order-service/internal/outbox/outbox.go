package outbox

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/config"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/rabbitmq"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/service"
	"gorm.io/gorm"
)

type Outbox struct {
	db                 *gorm.DB
	log                logger.Logger
	outboxEventService service.OutboxEventService
	rabbitmq           rabbitmq.RabbitMQService
	channels           []*amqp091.Channel
	config             *config.Outbox
}

type Opts struct {
	DB                 database.DatabaseService
	Log                logger.Logger
	OutboxEventService service.OutboxEventService
	RabbitMQ           rabbitmq.RabbitMQService
	Config             *config.Outbox
}

func NewOutbox(ctx context.Context, opts *Opts) Outbox {
	o := Outbox{
		db:                 opts.DB.DB(),
		log:                opts.Log,
		outboxEventService: opts.OutboxEventService,
		rabbitmq:           opts.RabbitMQ,
		config:             opts.Config,
	}

	hostname, _ := os.Hostname()
	go o.Start(ctx, hostname)

	return o
}

func (o *Outbox) Start(ctx context.Context, workerID string) {
	o.log.Info("Outbox worker started")

	if err := o.initChannels(o.config.MaxConcurrency); err != nil {
		o.log.Error("Failed to initialize channels", logger.Field{Key: "error", Value: err.Error()})
		return
	}
	defer o.closeChannels()

	eventsCh := make(chan *model.OutboxEvent, o.config.MaxConcurrency)
	wg := sync.WaitGroup{}

	for i := 0; i < o.config.MaxConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			o.processEvents(ctx, workerID, o.channels[i], eventsCh)
		}(i)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(eventsCh)
			wg.Wait()
			return

		case <-ticker.C:
			o.dispatchPendingEvents(ctx, workerID, eventsCh)
		}
	}
}

func (o *Outbox) initChannels(count int) error {
	for i := 0; i < count; i++ {
		ch, err := o.rabbitmq.NewChannel()
		if err != nil {
			return err
		}
		o.channels = append(o.channels, ch)
	}
	return nil
}

func (o *Outbox) closeChannels() {
	for _, ch := range o.channels {
		if ch != nil {
			_ = ch.Close()
		}
	}
}
