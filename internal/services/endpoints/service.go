package endpoints

import (
	"context"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"

	"github.com/google/uuid"
)

type EndpointRepo interface {
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Endpoint, error)
	GetAll(ctx context.Context, command dto.GetAllEndpointsCommand) ([]domain.Endpoint, int, error)
	Update(ctx context.Context, command dto.UpdateEndpointCommand) error
	Add(ctx context.Context, command dto.AddEndpointCommand) (*domain.Endpoint, error)
}

type Service struct {
	log  *slog.Logger
	repo EndpointRepo
}

func NewService(repo EndpointRepo, log *slog.Logger) *Service {
	return &Service{log: log, repo: repo}
}
