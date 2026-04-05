package model

import (
	"time"
)

type EventLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	EventType     string    `gorm:"not null;index" json:"event_type"`
	SourceService string    `gorm:"not null" json:"source_service"`
	Payload       string    `gorm:"type:text;not null" json:"payload"`
	CompanyID     uint      `gorm:"index" json:"company_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// PublishRequest is the JSON body for the publish endpoint.
type PublishRequest struct {
	EventType     string `json:"event_type" binding:"required"`
	SourceService string `json:"source_service" binding:"required"`
	Payload       string `json:"payload" binding:"required"`
	CompanyID     uint   `json:"company_id"`
}
