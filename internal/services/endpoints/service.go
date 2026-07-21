package endpoints

import (
	"context"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

type EndpointRepo interface {
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Endpoint, error)
	GetAll(ctx context.Context, command dto.GetAllEndpointsCommand) ([]domain.Endpoint, int, error)
	Update(ctx context.Context, command dto.UpdateEndpointCommand) error
	Add(ctx context.Context, command dto.AddEndpointCommand) (*dto.AddEndpointResult, error)
}

type Service struct {
	log  *slog.Logger
	repo EndpointRepo
}

func NewService(repo EndpointRepo, log *slog.Logger) *Service {
	return &Service{log: log, repo: repo}
}
