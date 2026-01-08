package model

import "time"

type Order struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Status    string    `gorm:"not null" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
