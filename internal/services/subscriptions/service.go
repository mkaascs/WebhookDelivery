package subscriptions

import (
	"context"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

type SubscriptionRepo interface {
	Add(ctx context.Context, command dto.AddSubscriptionCommand) ([]domain.Subscription, error)
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context, endpointID string) ([]domain.Subscription, error)
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
