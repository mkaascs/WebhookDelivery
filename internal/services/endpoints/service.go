package endpoints

import (
	"context"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

type Repo interface {
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Endpoint, error)
	GetAll(ctx context.Context, command dto.GetAllEndpointsCommand) ([]domain.Endpoint, int, error)
	Update(ctx context.Context, command dto.UpdateEndpointCommand) error
	AddEndpoint(ctx context.Context, command dto.RegisterEndpointCommand) (*dto.RegisterEndpointResult, error)
}

type Service struct {
	log  *slog.Logger
	repo Repo
}

func NewService(log *slog.Logger, repo Repo) *Service {
	return &Service{log: log, repo: repo}
}
