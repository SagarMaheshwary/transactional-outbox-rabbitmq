package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
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

func (o *orderService) Create(ctx context.Context, req *CreateOrder) (order *model.Order, err error) {
	tx := o.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	order = &model.Order{
		Status: "pending",
	}

	if err = tx.Create(order).Error; err != nil {
		return nil, err
	}

	outboxEvent := &model.OutboxEvent{
		ID:       uuid.NewString(),
		EventKey: "order.created",
		Payload: model.JSONB{
			"id":         order.ID,
			"product_id": req.ProductID,
			"quantity":   req.Quantity,
		},
		Status: model.OutboxEventStatusPending,
	}

	if err = o.outboxEventService.Create(ctx, tx, outboxEvent); err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	return order, nil
}
