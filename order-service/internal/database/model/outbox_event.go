package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type OutboxEvent struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	EventKey    string    `gorm:"not null" json:"event_key"`
	Payload     JSONB     `gorm:"type:jsonb;not null" json:"payload"`
	Status      string    `gorm:"not null" json:"status"`
	LockedAt    time.Time `json:"locked_at"`
	LockedBy    string    `json:"locked_by"`
	Traceparent string    `json:"traceparent"`
	CreatedAt   time.Time `json:"created_at"`
}

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

var (
	OutboxEventStatusPending    = "pending"
	OutboxEventStatusInProgress = "in_progress"
	OutboxEventStatusPublished  = "published"
	OutboxEventStatusFailed     = "failed"
)
