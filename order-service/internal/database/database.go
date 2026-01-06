package database

import (
	"context"
	"fmt"

	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/config"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseService interface {
	DB() *gorm.DB
	Close() error
	Health(ctx context.Context) error
}

type Database struct {
	db  *gorm.DB
	log logger.Logger
}

type Opts struct {
	Config *config.Database
	Log    logger.Logger
}

func NewDatabase(opts *Opts) (DatabaseService, error) {
	db, err := gorm.Open(postgres.Open(opts.Config.DSN))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %v", err)
	}

	opts.Log.Info("Database connected")

	return &Database{db: db, log: opts.Log}, nil
}

func (d *Database) DB() *gorm.DB {
	return d.db
}

func (d *Database) Close() error {
	if d == nil || d.db == nil {
		return fmt.Errorf("cannot close: database is not initialized")
	}

	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) Health(ctx context.Context) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}
