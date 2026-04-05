package model

import (
	"time"
)

type Webhook struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CompanyID uint      `gorm:"index;not null" json:"company_id"`
	URL       string    `gorm:"not null" json:"url"`
	Events    string    `gorm:"type:text;not null" json:"events"` // JSON array of event types
	Secret    string    `gorm:"not null" json:"secret"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WebhookCreate struct {
	CompanyID uint   `json:"company_id" binding:"required"`
	URL       string `json:"url" binding:"required,url"`
	Events    string `json:"events" binding:"required"` // JSON array string
	Secret    string `json:"secret"`
	IsActive  *bool  `json:"is_active"`
}

type WebhookUpdate struct {
	URL      *string `json:"url"`
	Events   *string `json:"events"`
	Secret   *string `json:"secret"`
	IsActive *bool   `json:"is_active"`
}
