package service

import (
	"context"

	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/database"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/logger"
	"gorm.io/gorm"
)

type ProcessedMessageService interface {
	TryInsert(ctx context.Context, tx *gorm.DB, row *model.ProcessedMessage) (bool, error)
}

type processedMessageService struct {
	db  *gorm.DB
	log logger.Logger
}

type ProcessedMessageOpts struct {
	DB  database.DatabaseService
	Log logger.Logger
}

func NewProcessedMessageService(opts *ProcessedMessageOpts) ProcessedMessageService {
	return &processedMessageService{
		db:  opts.DB.DB(),
		log: opts.Log,
	}
}

func (p *processedMessageService) TryInsert(
	ctx context.Context,
	tx *gorm.DB,
	row *model.ProcessedMessage,
) (bool, error) {
	query := `
		INSERT INTO processed_messages (message_id, processed_at)
		VALUES (?, ?)
		ON CONFLICT (message_id) DO NOTHING
	`

	res := tx.WithContext(ctx).Exec(
		query,
		row.MessageID,
		row.ProcessedAt,
	)

	if res.Error != nil {
		return false, res.Error
	}

	return res.RowsAffected > 0, nil
}
