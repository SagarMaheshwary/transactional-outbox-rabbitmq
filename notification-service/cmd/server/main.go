package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/config"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/database"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/rabbitmq"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := logger.NewZerologLogger("info", os.Stderr)

	cfg, err := config.NewConfig(".env")
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := database.NewDatabase(&database.Opts{
		Config: cfg.Database,
		Log:    log,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	processedMessageService := service.NewProcessedMessageService(&service.ProcessedMessageOpts{
		DB:  db,
		Log: log,
	})

	rmq, err := rabbitmq.NewRabbitMQ(ctx, &rabbitmq.Opts{
		Config:                  cfg.AMQP,
		Logger:                  log,
		ProcessedMessageService: processedMessageService,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	<-ctx.Done()

	log.Warn("Shutdown signal received, closing services!")

	if err := rmq.Close(); err != nil {
		log.Error("failed to close rabbitmq client", logger.Field{Key: "error", Value: err.Error()})
	}
	if err := db.Close(); err != nil {
		log.Error("failed to close database client", logger.Field{Key: "error", Value: err.Error()})
	}

	log.Info("Shutdown complete!")
}
