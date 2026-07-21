package deliveries

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"
)

func (s *Service) GetByID(ctx context.Context, id string) (*dto.GetDeliveryResult, error) {
	const fn = "services.deliveries.Service.GetByID"
	log := s.log.With(slog.String("fn", fn))

	delivery, err := s.repo.GetByID(ctx, id)
	if err != nil {
		const msg = "failed to get delivery by id"
		if utils.IsCtxError(err) || utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("id", id))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("id", id))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	return &dto.GetDeliveryResult{
		Delivery: *delivery,
	}, nil
}

func (s *Service) GetFromEvent(ctx context.Context, eventID string) ([]dto.GetDeliveryResult, error) {
	const fn = "services.deliveries.Service.GetFromEvent"
	log := s.log.With(slog.String("fn", fn))

	delivery, err := s.repo.GetFromEvent(ctx, eventID)
	if err != nil {
		const msg = "failed to get deliveries from event"
		if utils.IsCtxError(err) || utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("event_id", eventID))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("event_id", eventID))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	result := make([]dto.GetDeliveryResult, 0, len(delivery))
	for _, delivery := range delivery {
		result = append(result, dto.GetDeliveryResult{
			Delivery: delivery,
		})
	}

	return result, nil
}
