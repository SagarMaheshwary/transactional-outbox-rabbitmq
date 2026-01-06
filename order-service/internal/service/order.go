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
	Create(ctx context.Context, row *model.Order) (*model.Order, error)
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

func NewOrderService(opts *OrderServiceOpts) OrderService {
	return &orderService{
		db:                 opts.DB.DB(),
		log:                opts.Log,
		outboxEventService: opts.OutboxEventService,
	}
}

func (o *orderService) Create(ctx context.Context, row *model.Order) (*model.Order, error) {
	tx := o.db.Begin()

	if err := tx.WithContext(ctx).Create(row).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	outboxEvent := &model.OutboxEvent{
		EventKey: "order.created",
		Payload: model.JSONB{
			"id":      row.ID,
			"user_id": row.UserID,
			"amount":  row.Amount,
		},
		Status: model.OutboxEventStatusPending,
		ID:     uuid.NewString(),
	}
	if err := o.outboxEventService.Create(ctx, tx, outboxEvent); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return row, nil
}
