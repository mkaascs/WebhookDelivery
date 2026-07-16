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
			continue
		}

		if isSuccessCode(code) {
			successCount++
			s.handleSuccessRequest(ctx, delivery)
			continue
		}

		failedCount++
		s.handleFailedRequest(ctx, delivery)
	}

	log.Info("deliveries batch was processed",
		slog.Int("total", len(deliveries)),
		slog.Int("success", successCount),
		slog.Int("failed", failedCount))
}

func (s *Service) handleSuccessRequest(ctx context.Context, successDelivery dto.ClaimPendingResult) {
	err := s.deliveryRepo.UpdateStatus(ctx, dto.UpdateDeliveryStatusCommand{
		Status:      domain.StatusDelivered,
		Attempts:    successDelivery.Attempts,
		NextRetryAt: successDelivery.NextRetryAt,
	})

	if err != nil {
		s.log.Warn("failed to update status", sloglib.Error(err))
	}
}

func (s *Service) handleFailedRequest(ctx context.Context, failedDelivery dto.ClaimPendingResult) {
	err := s.deliveryRepo.UpdateStatus(ctx, dto.UpdateDeliveryStatusCommand{
		Status:      domain.StatusFailed,
		Attempts:    failedDelivery.Attempts,
		NextRetryAt: failedDelivery.NextRetryAt.Add(s.backoff(failedDelivery.Attempts)),
	})

	if err != nil {
		s.log.Warn("failed to update status", sloglib.Error(err))
	}
}

func (s *Service) backoff(attempts int) time.Duration {
	delay := s.cfg.BaseBackoff * (2 << (attempts - 1))
	if delay > float64(s.cfg.MaxBackoff) {
		delay = float64(s.cfg.MaxBackoff)
	}

	jitter := delay * 0.2 * (rand.Float64()*2 - 1)
	delay += jitter

	return time.Duration(delay)
}
