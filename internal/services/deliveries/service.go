package deliveries

import (
	"context"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

type RetryNotifier interface {
	Notify()
}

type DeliveryRepo interface {
	GetByID(ctx context.Context, id string) (*domain.Delivery, error)
	GetFromEvent(ctx context.Context, command dto.GetDeliveriesFromEventCommand) ([]domain.Delivery, int, error)
	UpdateStatus(ctx context.Context, command dto.UpdateDeliveryStatusCommand) error
}

type Service struct {
	log      *slog.Logger
	repo     DeliveryRepo
	notifier RetryNotifier
}

func NewService(repo DeliveryRepo, notifier RetryNotifier, log *slog.Logger) *Service {
	return &Service{repo: repo, notifier: notifier, log: log}
}
