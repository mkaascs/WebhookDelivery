package endpoints

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	"webhook-delivery/internal/lib/crypto"
	sloglib "webhook-delivery/internal/lib/logging/slog"

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

func (s *Service) Register(ctx context.Context, command dto.RegisterEndpointCommand) (*domain.Endpoint, error) {
	const fn = "services.endpoints.Service.Register"
	log := s.log.With(slog.String("fn", fn))

	secret := crypto.NewSecret()

	endpoint, err := s.repo.Add(ctx, dto.AddEndpointCommand{
		URL:         command.URL,
		EventTypes:  command.EventTypes,
		Description: command.Description,
		Secret:      secret,
	})

	if err != nil {
		const msg = "failed to register endpoint"
		s.logError(log, msg, err)
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("new endpoint was created successfully", slog.String("endpoint_id", endpoint.ID.String()))

	return endpoint, nil
}

func (s *Service) Update(ctx context.Context, command dto.UpdateEndpointCommand) error {
	const fn = "services.endpoints.Service.Update"
	log := s.log.With(slog.String("fn", fn))

	if err := s.repo.Update(ctx, command); err != nil {
		const msg = "failed to update endpoint"
		s.logError(log, msg, err, slog.String("endpoint_id", command.ID.String()))
		return fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("endpoint was updated successfully", slog.String("endpoint_id", command.ID.String()))

	return nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Endpoint, error) {
	const fn = "services.endpoints.Service.GetByID"
	log := s.log.With(slog.String("fn", fn))

	endpoint, err := s.repo.GetByID(ctx, id)
	if err != nil {
		const msg = "failed to get endpoint"
		s.logError(log, msg, err, slog.String("endpoint_id", id.String()))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("endpoint info was sent successfully", slog.String("endpoint_id", id.String()))

	return endpoint, nil
}

func (s *Service) GetAll(ctx context.Context, command dto.GetAllEndpointsCommand) ([]domain.Endpoint, int, error) {
	const fn = "services.endpoints.Service.GetAll"
	log := s.log.With(slog.String("fn", fn))

	endpoints, total, err := s.repo.GetAll(ctx, command)
	if err != nil {
		const msg = "failed to get all endpoints"
		s.logError(log, msg, err)
		return nil, 0, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("endpoints info were sent successfully",
		slog.Int("total", total),
		slog.Int("limit", command.Limit),
		slog.Int("page", command.Page))

	return endpoints, total, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	const fn = "services.endpoints.Service.Delete"
	log := s.log.With(slog.String("fn", fn))

	if err := s.repo.Delete(ctx, id); err != nil {
		const msg = "failed to delete endpoint"
		s.logError(log, msg, err, slog.String("endpoint_id", id.String()))
		return fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("endpoint successfully deleted", slog.String("endpoint_id", id.String()))

	return nil
}

func (s *Service) logError(log *slog.Logger, msg string, err error, args ...any) {
	if domain.IsDomainError(err) || domain.IsCtxError(err) {
		log.With(args...).Info(msg, sloglib.Error(err))
		return
	}

	log.With(args...).Error(msg, sloglib.Error(err))
}
