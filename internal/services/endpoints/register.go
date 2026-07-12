package endpoints

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"
)

func (s *Service) Register(ctx context.Context, command dto.RegisterEndpointCommand) (*dto.RegisterEndpointResult, error) {
	const fn = "services.endpoints.Service.Register"
	log := s.log.With(slog.String("fn", fn))

	result, err := s.repo.AddEndpoint(ctx, command)
	if err != nil {
		const msg = "failed to register endpoint"
		if utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("new endpoint was created successfully", slog.String("id", result.ID))

	return result, nil
}
