package deliveries

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"
)

func (s *Service) Retry(ctx context.Context, id string) error {
	const fn = "services.deliveries.Service.Retry"
	log := s.log.With(slog.String("fn", fn))

	cmd := dto.UpdateDeliveryStatusCommand{
		ID:               id,
		Status:           domain.StatusPending,
		Attempts:         0,
		NextRetryAt:      time.Now(),
		LastError:        nil,
		LastResponseCode: nil,
	}

	if err := s.repo.UpdateStatus(ctx, cmd); err != nil {
		const msg = "failed to retry delivery"
		if utils.IsCtxError(err) || utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("id", id))
			return fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("id", id))
		return fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	s.notifier.Notify()

	log.Info("delivery was retried", slog.String("id", id))

	return nil
}
