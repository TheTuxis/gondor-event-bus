package repository

import (
	"github.com/TheTuxis/gondor-event-bus/internal/model"
	"gorm.io/gorm"
)

type DeadLetterRepository struct {
	db *gorm.DB
}

func NewDeadLetterRepository(db *gorm.DB) *DeadLetterRepository {
	return &DeadLetterRepository{db: db}
}

func (r *DeadLetterRepository) FindByID(id uint) (*model.DeadLetterMessage, error) {
	var msg model.DeadLetterMessage
	if err := r.db.First(&msg, id).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *DeadLetterRepository) List(params model.ListParams) ([]model.DeadLetterMessage, int64, error) {
	var messages []model.DeadLetterMessage
	var total int64

	query := r.db.Model(&model.DeadLetterMessage{})

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("subject ILIKE ? OR error_message ILIKE ?", search, search)
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
		Find(&messages).Error

	return messages, total, err
}

func (r *DeadLetterRepository) Create(msg *model.DeadLetterMessage) error {
	return r.db.Create(msg).Error
}

func (r *DeadLetterRepository) Update(msg *model.DeadLetterMessage) error {
	return r.db.Save(msg).Error
}

func (r *DeadLetterRepository) Delete(id uint) error {
	return r.db.Delete(&model.DeadLetterMessage{}, id).Error
}
