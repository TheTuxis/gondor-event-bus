package service

import (
	"errors"

	"go.uber.org/zap"

	"github.com/TheTuxis/gondor-event-bus/internal/model"
	"github.com/TheTuxis/gondor-event-bus/internal/repository"
)

var (
	ErrDLQMessageNotFound = errors.New("dead letter message not found")
)

type DeadLetterService struct {
	dlqRepo *repository.DeadLetterRepository
	logger  *zap.Logger
}

func NewDeadLetterService(dlqRepo *repository.DeadLetterRepository, logger *zap.Logger) *DeadLetterService {
	return &DeadLetterService{dlqRepo: dlqRepo, logger: logger}
}

func (s *DeadLetterService) List(params model.ListParams) (*model.PaginatedResult, error) {
	messages, total, err := s.dlqRepo.List(params)
	if err != nil {
		return nil, err
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return &model.PaginatedResult{
		Data: messages,
		Pagination: model.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	}, nil
}

func (s *DeadLetterService) Retry(id uint) (*model.DeadLetterMessage, error) {
	msg, err := s.dlqRepo.FindByID(id)
	if err != nil {
		return nil, ErrDLQMessageNotFound
	}

	if msg.Status == "exhausted" {
		return nil, errors.New("message has exhausted all retries")
	}

	msg.Status = "retrying"
	msg.RetryCount++

	if err := s.dlqRepo.Update(msg); err != nil {
		return nil, err
	}

	s.logger.Info("dead letter message queued for retry",
		zap.Uint("id", id),
		zap.Int("retry_count", msg.RetryCount),
	)

	return msg, nil
}

func (s *DeadLetterService) Delete(id uint) error {
	if _, err := s.dlqRepo.FindByID(id); err != nil {
		return ErrDLQMessageNotFound
	}
	return s.dlqRepo.Delete(id)
}
