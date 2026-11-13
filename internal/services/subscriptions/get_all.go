package subscriptions

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"

	"github.com/google/uuid"
)

func (s *Service) GetAll(ctx context.Context, endpointID uuid.UUID) ([]domain.Subscription, error) {
	const fn = "services.subscriptions.Service.GetAll"
	log := s.log.With(slog.String("fn", fn))

	subs, err := s.repo.GetAll(ctx, endpointID)
	if err != nil {
		const msg = "failed to get all subscriptions of endpoint"
		if utils.IsCtxError(err) || utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("endpoint_id", endpointID.String()))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("endpoint_id", endpointID.String()))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("subscriptions info were sent successfully", slog.String("endpoint_id", endpointID.String()))
	return subs, nil
}
