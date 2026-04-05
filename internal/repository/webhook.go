package repository

import (
	"github.com/TheTuxis/gondor-event-bus/internal/model"
	"gorm.io/gorm"
)

type WebhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) FindByID(id uint) (*model.Webhook, error) {
	var webhook model.Webhook
	if err := r.db.First(&webhook, id).Error; err != nil {
		return nil, err
	}
	return &webhook, nil
}

func (r *WebhookRepository) List(params model.ListParams) ([]model.Webhook, int64, error) {
	var webhooks []model.Webhook
	var total int64

	query := r.db.Model(&model.Webhook{})

	if params.CompanyID != nil {
		query = query.Where("company_id = ?", *params.CompanyID)
	}
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("url ILIKE ?", search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	sortBy := "id"
	if params.SortBy != "" {
		sortBy = params.SortBy
	}
	sortOrder := "asc"
	if params.SortOrder == "desc" {
		sortOrder = "desc"
	}

	offset := (params.Page - 1) * params.PageSize
	err := query.
		Order(sortBy + " " + sortOrder).
		Offset(offset).Limit(params.PageSize).
		Find(&webhooks).Error

	return webhooks, total, err
}

func (r *WebhookRepository) Create(webhook *model.Webhook) error {
	return r.db.Create(webhook).Error
}

func (r *WebhookRepository) Update(webhook *model.Webhook) error {
	return r.db.Save(webhook).Error
}

func (r *WebhookRepository) Delete(id uint) error {
	return r.db.Delete(&model.Webhook{}, id).Error
}

// FindActiveByEventType returns all active webhooks whose events JSON array contains the given event type.
func (r *WebhookRepository) FindActiveByEventType(eventType string) ([]model.Webhook, error) {
	var webhooks []model.Webhook
	err := r.db.Where("is_active = ? AND events LIKE ?", true, "%"+eventType+"%").Find(&webhooks).Error
	return webhooks, err
}
