package endpoints

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"
)

func (s *Service) GetByID(ctx context.Context, id string) (*dto.GetEndpointResult, error) {
	const fn = "services.endpoints.Service.GetByID"
	log := s.log.With(slog.String("fn", fn))

	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		const msg = "failed to get endpoint"
		if utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("id", id))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("id", id))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("endpoint info was sent successfully", slog.String("id", id))

	return result, nil
}

func (s *Service) GetAll(ctx context.Context, command dto.GetAllEndpointsCommand) (*dto.GetAllEndpointsResult, error) {
	const fn = "services.endpoints.Service.GetAll"
	log := s.log.With(slog.String("fn", fn))

	result, err := s.repo.GetAll(ctx, command)
	if err != nil {
		const msg = "failed to get all endpoints"
		if utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("endpoints info were sent successfully",
		slog.Int("total", result.Total),
		slog.Int("limit", command.Limit),
		slog.Int("page", command.Page))

	return result, nil
}
