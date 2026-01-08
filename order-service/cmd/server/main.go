package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/config"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database"
	httpserver "github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/http"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/outbox"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/rabbitmq"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/service"
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

	rmq, err := rabbitmq.NewRabbitMQ(ctx, &rabbitmq.Opts{
		Config: cfg.AMQP,
		Logger: log,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	outboxEventService := service.NewOutboxEventService(&service.OutboxEventServiceOpts{
		DB:  db,
		Log: log,
	})
	orderService := service.NewOrderService(&service.OrderServiceOpts{
		DB:                 db,
		Log:                log,
		OutboxEventService: outboxEventService,
	})

	outbox.NewOutbox(ctx, &outbox.Opts{
		DB:                 db,
		Log:                log,
		OutboxEventService: outboxEventService,
		RabbitMQ:           rmq,
		Config:             cfg.Outbox,
	})

	httpServer := httpserver.NewServer(cfg.HTTPServer.URL, &httpserver.Opts{
		OrderService: orderService,
		Log:          log,
	})
	go func() {
		err = httpServer.Serve()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			stop()
		}
	}()

	<-ctx.Done()

	log.Warn("Shutdown signal received, closing services!")

	if err := rmq.Close(); err != nil {
		log.Error("failed to close rabbitmq client", logger.Field{Key: "error", Value: err.Error()})
	}
	if err := db.Close(); err != nil {
		log.Error("failed to close database client", logger.Field{Key: "error", Value: err.Error()})
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 3*time.Second)
	if err := httpServer.Server.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to close http server", logger.Field{Key: "error", Value: err.Error()})
	}
	cancelShutdown()

	log.Info("Shutdown complete!")
}
