package workers

import (
	"context"
	"log/slog"
	"math/rand"
	"time"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"
)

func (s *Service) processBatch(ctx context.Context) {
	const fn = "services.workers.Service.processBatch"
	log := s.log.With(slog.String("fn", fn))

	deliveries, err := s.deliveryRepo.ClaimPending(ctx, s.cfg.BatchSize)
	if err != nil {
		const msg = "failed to claim pending deliveries"
		if utils.IsCtxError(err) {
			log.Info(msg, sloglib.Error(err))
			return
		}

		log.Warn(msg, sloglib.Error(err))
		return
	}

	if len(deliveries) == 0 {
		return
	}

	s.Notify()

	successCount, failedCount := 0, 0
	for _, delivery := range deliveries {
		delivery.Attempts++
		code, err := sendPostRequest(delivery.URL, delivery.Payload, delivery.Secret)
		if err != nil {
			log.Warn("failed to send post request", sloglib.Error(err), slog.String("url", delivery.URL))
		}

		cmd := dto.UpdateDeliveryStatusCommand{
			ID:          delivery.ID,
			Status:      domain.StatusPending,
			Attempts:    delivery.Attempts,
			NextRetryAt: delivery.NextRetryAt,
		}

		if isSuccessCode(code) {
			successCount++
			cmd.Status = domain.StatusDelivered
		} else if delivery.Attempts >= delivery.MaxAttempts {
			failedCount++
			cmd.Status = domain.StatusFailed
		}

		if cmd.Status == domain.StatusPending {
			cmd.NextRetryAt = time.Now().Add(s.backoff(delivery.Attempts))
		}

		if err = s.deliveryRepo.UpdateStatus(ctx, cmd); err != nil {
			s.log.Warn("failed to update status", sloglib.Error(err))
		}
	}

	log.Info("deliveries batch was processed",
		slog.Int("total", len(deliveries)),
		slog.Int("success", successCount),
		slog.Int("failed", failedCount))
}

func (s *Service) backoff(attempts int) time.Duration {
	multiplier := 1 << (attempts - 1)
	delay := float64(s.cfg.BaseBackoff) * float64(multiplier)
	if delay > float64(s.cfg.MaxBackoff) {
		delay = float64(s.cfg.MaxBackoff)
	}

	jitter := delay * 0.2 * (rand.Float64()*2 - 1)
	delay += jitter

	return time.Duration(delay)
}
