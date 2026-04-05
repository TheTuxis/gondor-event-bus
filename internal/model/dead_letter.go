package model

import (
	"time"
)

type DeadLetterMessage struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Subject      string     `gorm:"not null;index" json:"subject"`
	Data         string     `gorm:"type:text;not null" json:"data"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	RetryCount   int        `gorm:"default:0" json:"retry_count"`
	MaxRetries   int        `gorm:"default:3" json:"max_retries"`
	NextRetryAt  *time.Time `json:"next_retry_at"`
	Status       string     `gorm:"default:pending;index" json:"status"` // pending, retrying, exhausted
	CreatedAt    time.Time  `json:"created_at"`
}
