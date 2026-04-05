package repository

import (
	"github.com/TheTuxis/gondor-event-bus/internal/model"
	"gorm.io/gorm"
)

type EventLogRepository struct {
	db *gorm.DB
}

func NewEventLogRepository(db *gorm.DB) *EventLogRepository {
	return &EventLogRepository{db: db}
}

func (r *EventLogRepository) Create(log *model.EventLog) error {
	return r.db.Create(log).Error
}

func (r *EventLogRepository) List(params model.ListParams) ([]model.EventLog, int64, error) {
	var logs []model.EventLog
	var total int64

	query := r.db.Model(&model.EventLog{})

	if params.CompanyID != nil {
		query = query.Where("company_id = ?", *params.CompanyID)
	}
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("event_type ILIKE ? OR source_service ILIKE ?", search, search)
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

	sortBy := "created_at"
	sortOrder := "desc"
	if params.SortBy != "" {
		sortBy = params.SortBy
	}
	if params.SortOrder == "asc" {
		sortOrder = "asc"
	}

	offset := (params.Page - 1) * params.PageSize
	err := query.
		Order(sortBy + " " + sortOrder).
		Offset(offset).Limit(params.PageSize).
		Find(&logs).Error

	return logs, total, err
}
