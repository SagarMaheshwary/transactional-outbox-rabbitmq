package model

import "time"

type ProcessedMessage struct {
	MessageID   string    `gorm:"primaryKey" json:"message_id"`
	ProcessedAt time.Time `json:"processed_at"`
}
