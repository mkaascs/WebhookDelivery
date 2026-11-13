package subscriptions

import (
	"context"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"

	"github.com/google/uuid"
)

type SubscriptionRepo interface {
	Add(ctx context.Context, command dto.AddSubscriptionCommand) ([]domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context, endpointID uuid.UUID) ([]domain.Subscription, error)
}

type Service struct {
	log  *slog.Logger
	repo SubscriptionRepo
}

func NewService(repo SubscriptionRepo, log *slog.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}
