package service

import (
	"context"

	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"gorm.io/gorm"
)

type OutboxEventService interface {
	Create(ctx context.Context, tx *gorm.DB, row *model.OutboxEvent) error
	ClaimEvents(ctx context.Context, workerID string, limit int) ([]*model.OutboxEvent, error)
	UpdateStateIfInProgress(ctx context.Context, eventID string, update map[string]interface{}) (bool, error)
	CountBacklog(ctx context.Context) (int64, error)
	CountWaitingRetry(ctx context.Context) (int64, error)
}

type outboxEventService struct {
	db  *gorm.DB
	log logger.Logger
}

type OutboxEventServiceOpts struct {
	DB  database.DatabaseService
	Log logger.Logger
}

func NewOutboxEventService(opts *OutboxEventServiceOpts) OutboxEventService {
	return &outboxEventService{
		db:  opts.DB.DB(),
		log: opts.Log,
	}
}

func (o *outboxEventService) Create(ctx context.Context, tx *gorm.DB, row *model.OutboxEvent) error {
	return tx.WithContext(ctx).Save(row).Error
}

func (o *outboxEventService) UpdateStateIfInProgress(
	ctx context.Context,
	eventID string,
	update map[string]interface{},
) (bool, error) {

	result := o.db.WithContext(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ? AND status = ?", eventID, model.OutboxEventStatusInProgress).
		UpdateColumns(update)

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}

func (o *outboxEventService) ClaimEvents(
	ctx context.Context,
	workerID string,
	limit int,
) ([]*model.OutboxEvent, error) {
	var events []*model.OutboxEvent

	query := `
		UPDATE outbox_events
		SET
			status = ?,
			locked_at = NOW(),
			locked_by = ?
		WHERE id IN (
				SELECT id
				FROM outbox_events
				WHERE
					(
						status = ?
						OR (
							status = ?
							AND locked_at < NOW() - INTERVAL '30 seconds'
						)
					)
					AND (
						next_retry_at IS NULL
						OR next_retry_at <= NOW()
					)
				ORDER BY created_at
				LIMIT ?
				FOR UPDATE SKIP LOCKED
		)
		RETURNING *`

	err := o.db.WithContext(ctx).
		Raw(query,
			model.OutboxEventStatusInProgress,
			workerID,
			model.OutboxEventStatusPending,
			model.OutboxEventStatusInProgress,
			limit,
		).
		Scan(&events).Error

	return events, err
}

func (o *outboxEventService) CountBacklog(ctx context.Context) (int64, error) {
	var count int64

	err := o.db.WithContext(ctx).
		Model(&model.OutboxEvent{}).
		Where(`
			(
				status = ?
				OR (
					status = ?
					AND locked_at < NOW() - INTERVAL '30 seconds'
				)
			)
			AND (
				next_retry_at IS NULL
				OR next_retry_at <= NOW()
			)`,
			model.OutboxEventStatusPending,
			model.OutboxEventStatusInProgress,
		).
		Count(&count).Error

	return count, err
}

func (o *outboxEventService) CountWaitingRetry(ctx context.Context) (int64, error) {
	var count int64

	err := o.db.WithContext(ctx).
		Model(&model.OutboxEvent{}).
		Where(`
				status = ? 
			AND 
				next_retry_at IS NOT NULL 
			AND 
				next_retry_at > NOW()
		`,
			model.OutboxEventStatusPending,
		).
		Count(&count).Error

	return count, err
}
