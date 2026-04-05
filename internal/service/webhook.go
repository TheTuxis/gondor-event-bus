package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"go.uber.org/zap"

	"github.com/TheTuxis/gondor-event-bus/internal/model"
	"github.com/TheTuxis/gondor-event-bus/internal/repository"
)

var (
	ErrWebhookNotFound = errors.New("webhook not found")
)

type WebhookService struct {
	webhookRepo *repository.WebhookRepository
	logger      *zap.Logger
}

func NewWebhookService(webhookRepo *repository.WebhookRepository, logger *zap.Logger) *WebhookService {
	return &WebhookService{webhookRepo: webhookRepo, logger: logger}
}

func (s *WebhookService) List(params model.ListParams) (*model.PaginatedResult, error) {
	webhooks, total, err := s.webhookRepo.List(params)
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
		Data: webhooks,
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

func (s *WebhookService) GetByID(id uint) (*model.Webhook, error) {
	webhook, err := s.webhookRepo.FindByID(id)
	if err != nil {
		return nil, ErrWebhookNotFound
	}
	return webhook, nil
}

func (s *WebhookService) Create(input model.WebhookCreate) (*model.Webhook, error) {
	secret := input.Secret
	if secret == "" {
		secret = generateSecret()
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	webhook := &model.Webhook{
		CompanyID: input.CompanyID,
		URL:       input.URL,
		Events:    input.Events,
		Secret:    secret,
		IsActive:  isActive,
	}

	if err := s.webhookRepo.Create(webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

func (s *WebhookService) Update(id uint, input model.WebhookUpdate) (*model.Webhook, error) {
	webhook, err := s.webhookRepo.FindByID(id)
	if err != nil {
		return nil, ErrWebhookNotFound
	}

	if input.URL != nil {
		webhook.URL = *input.URL
	}
	if input.Events != nil {
		webhook.Events = *input.Events
	}
	if input.Secret != nil {
		webhook.Secret = *input.Secret
	}
	if input.IsActive != nil {
		webhook.IsActive = *input.IsActive
	}

	if err := s.webhookRepo.Update(webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

func (s *WebhookService) Delete(id uint) error {
	if _, err := s.webhookRepo.FindByID(id); err != nil {
		return ErrWebhookNotFound
	}
	return s.webhookRepo.Delete(id)
}

// Test sends a test event payload to the webhook URL (placeholder — no HTTP call yet).
func (s *WebhookService) Test(id uint) (*model.Webhook, error) {
	webhook, err := s.webhookRepo.FindByID(id)
	if err != nil {
		return nil, ErrWebhookNotFound
	}
	s.logger.Info("test event sent to webhook", zap.Uint("webhook_id", id), zap.String("url", webhook.URL))
	return webhook, nil
}

func generateSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
