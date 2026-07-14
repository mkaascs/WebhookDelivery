package subscriptions

import (
	"context"
	"fmt"
	"log/slog"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"
)

func (s *Service) Delete(ctx context.Context, id string) error {
	const fn = "services.subscriptions.Service.Delete"
	log := s.log.With(slog.String("fn", fn))

	if err := s.repo.Delete(ctx, id); err != nil {
		const msg = "failed to delete subscription"
		if utils.IsDomainError(err) || utils.IsCtxError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("id", id))
			return fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("id", id))
		return fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("subscription was deleted successfully", slog.String("id", id))

	return nil
}
