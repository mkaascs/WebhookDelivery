package subscriptions

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"
)

func (s *Service) GetAll(ctx context.Context, endpointID string) ([]dto.GetSubscriptionResult, error) {
	const fn = "services.subscriptions.Service.GetAll"
	log := s.log.With(slog.String("fn", fn))

	subs, err := s.repo.GetAll(ctx, endpointID)
	if err != nil {
		const msg = "failed to get all subscriptions of endpoint"
		if utils.IsCtxError(err) || utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("endpoint_id", endpointID))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("endpoint_id", endpointID))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	results := make([]dto.GetSubscriptionResult, 0, len(subs))
	for _, sub := range subs {
		results = append(results, dto.GetSubscriptionResult{
			ID:         sub.ID,
			EndpointID: sub.EndpointID,
			EventType:  sub.EventType,
			CreatedAt:  sub.CreatedAt,
		})
	}

	log.Info("subscriptions info were sent successfully", slog.String("endpoint_id", endpointID))
	return results, nil
}
