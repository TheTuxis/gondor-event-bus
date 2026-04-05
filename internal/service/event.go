package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	natspkg "github.com/TheTuxis/gondor-event-bus/internal/nats"
	"github.com/TheTuxis/gondor-event-bus/internal/model"
	"github.com/TheTuxis/gondor-event-bus/internal/repository"
)

var (
	ErrEventLogNotFound = errors.New("event log not found")
)

type EventService struct {
	eventLogRepo *repository.EventLogRepository
	natsClient   *natspkg.Client
	logger       *zap.Logger
}

func NewEventService(eventLogRepo *repository.EventLogRepository, natsClient *natspkg.Client, logger *zap.Logger) *EventService {
	return &EventService{
		eventLogRepo: eventLogRepo,
		natsClient:   natsClient,
		logger:       logger,
	}
}

func (s *EventService) Publish(ctx context.Context, input model.PublishRequest) (*model.EventLog, error) {
	// Publish to NATS
	if s.natsClient != nil {
		if err := s.natsClient.Publish(ctx, input.EventType, []byte(input.Payload)); err != nil {
			s.logger.Error("failed to publish to NATS", zap.Error(err), zap.String("event_type", input.EventType))
			// Continue to log even if NATS publish fails
		}
	}

	// Persist event log
	eventLog := &model.EventLog{
		EventType:     input.EventType,
		SourceService: input.SourceService,
		Payload:       input.Payload,
		CompanyID:     input.CompanyID,
	}

	if err := s.eventLogRepo.Create(eventLog); err != nil {
		return nil, err
	}

	s.logger.Info("event published",
		zap.String("event_type", input.EventType),
		zap.String("source_service", input.SourceService),
	)

	return eventLog, nil
}

func (s *EventService) ListLogs(params model.ListParams) (*model.PaginatedResult, error) {
	logs, total, err := s.eventLogRepo.List(params)
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
		Data: logs,
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
