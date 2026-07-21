package deliveries

import (
	"context"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

type DeliveryRepo interface {
	GetByID(ctx context.Context, id string) (*domain.Delivery, error)
	GetFromEvent(ctx context.Context, eventID string) ([]domain.Delivery, error)
	UpdateStatus(ctx context.Context, command dto.UpdateDeliveryStatusCommand) error
}

type Service struct {
	log  *slog.Logger
	repo DeliveryRepo
}

func NewService(repo DeliveryRepo, log *slog.Logger) *Service {
	return &Service{repo: repo, log: log}
}
