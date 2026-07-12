package endpoints

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"
)

func (s *Service) Update(ctx context.Context, command dto.UpdateEndpointCommand) error {
	const fn = "services.endpoints.Service.Update"
	log := s.log.With(slog.String("fn", fn))

	if err := s.repo.Update(ctx, command); err != nil {
		const msg = "failed to update endpoint"
		if utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err))
			return fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
