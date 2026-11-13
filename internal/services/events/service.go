package events

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/google/uuid"
)

type EventRepo interface {
	Add(ctx context.Context, command dto.PublishEventCommand) (*domain.Event, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
}

type DeliveryRepo interface {
	CreateForEvent(ctx context.Context, eventID uuid.UUID, eventType string) (int, error)
}

type Service struct {
	log          *slog.Logger
	eventRepo    EventRepo
	deliveryRepo DeliveryRepo
}

func NewService(eventRepo EventRepo, deliveryRepo DeliveryRepo, log *slog.Logger) *Service {
	return &Service{
		log:          log,
		eventRepo:    eventRepo,
		deliveryRepo: deliveryRepo,
	}
}

func (s *Service) Publish(ctx context.Context, command dto.PublishEventCommand) (*domain.Event, int, error) {
	const fn = "services.events.Service.Publish"
	log := s.log.With(slog.String("fn", fn))

	event, err := s.eventRepo.Add(ctx, command)
	if err != nil {
		const msg = "failed to add event to repo"
		s.logError(log, msg, err)
		return nil, 0, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	count, err := s.deliveryRepo.CreateForEvent(ctx, event.ID, event.Type)
	if err != nil {
		const msg = "failed to create deliveries for event"
		s.logError(log, msg, err, slog.String("event_id", event.ID.String()))
		return nil, 0, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("event was published successfully", slog.String("event_id", event.ID.String()), slog.String("event_type", event.Type))

	return event, count, nil
}

func (s *Service) Get(ctx context.Context, eventID uuid.UUID) (*domain.Event, error) {
	const fn = "services.events.Service.Get"
	log := s.log.With(slog.String("fn", fn))

	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		const msg = "failed to get event by id"
		s.logError(log, msg, err, slog.String("event_id", eventID.String()))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("event info was sent successfully", slog.String("event_id", event.ID.String()))

	return event, nil
}

func (s *Service) logError(log *slog.Logger, msg string, err error, args ...any) {
	if domain.IsDomainError(err) || domain.IsCtxError(err) {
		log.With(args...).Info(msg, sloglib.Error(err))
		return
	}

	log.With(args...).Error(msg, sloglib.Error(err))
}
