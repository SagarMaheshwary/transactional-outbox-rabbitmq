package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/observability/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"gorm.io/gorm"
)

type OrderService interface {
	Create(ctx context.Context, req *CreateOrder) (*model.Order, error)
}

type orderService struct {
	db                 *gorm.DB
	log                logger.Logger
	outboxEventService OutboxEventService
}

type OrderServiceOpts struct {
	DB                 database.DatabaseService
	Log                logger.Logger
	OutboxEventService OutboxEventService
}

type CreateOrder struct {
	ProductID string
	Quantity  int
}

func NewOrderService(opts *OrderServiceOpts) OrderService {
	return &orderService{
		db:                 opts.DB.DB(),
		log:                opts.Log,
		outboxEventService: opts.OutboxEventService,
	}
}

func (o *orderService) Create(
	ctx context.Context,
	req *CreateOrder,
) (*model.Order, error) {
	ctx, span := tracing.Tracer.Start(ctx, "OrderService.Create")
	defer span.End()

	var order *model.Order

	err := withTransaction(ctx, o.db, func(tx *gorm.DB) error {
		order = &model.Order{
			Status: "pending",
		}

		if err := tx.Create(order).Error; err != nil {
			return err
		}

		carrier := propagation.MapCarrier{}
		otel.GetTextMapPropagator().Inject(ctx, carrier)

		outboxEvent := &model.OutboxEvent{
			ID:       uuid.NewString(),
			EventKey: "order.created",
			Payload: model.JSONB{
				"id":         order.ID,
				"product_id": req.ProductID,
				"quantity":   req.Quantity,
			},
			Status:      model.OutboxEventStatusPending,
			Traceparent: carrier["traceparent"],
		}

		if err := o.outboxEventService.Create(ctx, tx, outboxEvent); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return order, nil
}
