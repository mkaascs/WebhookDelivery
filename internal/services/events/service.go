package events

import (
	"context"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

type EventNotifier interface {
	Notify()
}

type EventRepo interface {
	Add(ctx context.Context, command dto.PublishEventCommand) (*domain.Event, error)
	GetByID(ctx context.Context, id string) (*domain.Event, error)
}

type DeliveryRepo interface {
	CreateForEvent(ctx context.Context, eventID string, eventType string) (int, error)
}

type Service struct {
	log          *slog.Logger
	notifier     EventNotifier
	eventRepo    EventRepo
	deliveryRepo DeliveryRepo
}

func NewService(eventRepo EventRepo, deliveryRepo DeliveryRepo, notifier EventNotifier, log *slog.Logger) *Service {
	return &Service{
		log:          log,
		notifier:     notifier,
		eventRepo:    eventRepo,
		deliveryRepo: deliveryRepo,
	}
}
