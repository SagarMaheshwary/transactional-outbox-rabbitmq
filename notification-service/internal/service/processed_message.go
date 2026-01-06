package service

import (
	"context"

	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/database"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/database/model"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (p *processedMessageService) TryInsert(ctx context.Context, tx *gorm.DB, row *model.ProcessedMessage) (bool, error) {
	if tx == nil {
		tx = p.db
	}

	res := tx.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}
