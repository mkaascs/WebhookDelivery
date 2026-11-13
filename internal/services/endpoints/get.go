package endpoints

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"

	"github.com/google/uuid"
)

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Endpoint, error) {
	const fn = "services.endpoints.Service.GetByID"
	log := s.log.With(slog.String("fn", fn))

	endpoint, err := s.repo.GetByID(ctx, id)
	if err != nil {
		const msg = "failed to get endpoint"
		if utils.IsDomainError(err) || utils.IsCtxError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("id", id.String()))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("id", id.String()))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("endpoint info was sent successfully", slog.String("id", id.String()))

	return endpoint, nil
}

func (s *Service) GetAll(ctx context.Context, command dto.GetAllEndpointsCommand) ([]domain.Endpoint, int, error) {
	const fn = "services.endpoints.Service.GetAll"
	log := s.log.With(slog.String("fn", fn))

	endpoints, total, err := s.repo.GetAll(ctx, command)
	if err != nil {
		const msg = "failed to get all endpoints"
		if utils.IsDomainError(err) || utils.IsCtxError(err) {
			log.Info(msg, sloglib.Error(err))
			return nil, 0, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err))
		return nil, 0, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("endpoints info were sent successfully",
		slog.Int("total", total),
		slog.Int("limit", command.Limit),
		slog.Int("page", command.Page))

	return endpoints, total, nil
}
