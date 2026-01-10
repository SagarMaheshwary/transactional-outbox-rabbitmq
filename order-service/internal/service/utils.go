package service

import (
	"context"

	"gorm.io/gorm"
)

func withTransaction(
	ctx context.Context,
	db *gorm.DB,
	fn func(tx *gorm.DB) error,
) (err error) {
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
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

	err = fn(tx)
	if err != nil {
		return err
	}

	return tx.Commit().Error
}
