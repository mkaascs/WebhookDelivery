package workers

import (
	"context"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

func (s *Service) processBatch(ctx context.Context) {
	const fn = "services.workers.Service.processBatch"
	log := s.log.With(slog.String("fn", fn))

	deliveries, err := s.deliveryRepo.ClaimPending(ctx, s.cfg.BatchSize)
	if err != nil {
		handleRepoErr(log, "failed to claim pending delivery", err)
		return
	}

	if len(deliveries) == 0 {
		return
	}

	s.Notify()

	successCount, failedCount := 0, 0
	for _, delivery := range deliveries {
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
		Attempts:    successDelivery.Attempts + 1,
		NextRetryAt: successDelivery.NextRetryAt,
	})

	if err != nil {
		s.log.Warn("failed to update status", sloglib.Error(err))
	}
}

func (s *Service) handleFailedRequest(ctx context.Context, failedDelivery dto.ClaimPendingResult) {
	err := s.deliveryRepo.UpdateStatus(ctx, dto.UpdateDeliveryStatusCommand{
		Status:      domain.StatusFailed,
		Attempts:    failedDelivery.Attempts + 1,
		NextRetryAt: failedDelivery.NextRetryAt.Add(backoff(failedDelivery.Attempts)),
	})

	if err != nil {
		s.log.Warn("failed to update status", sloglib.Error(err))
	}
}
