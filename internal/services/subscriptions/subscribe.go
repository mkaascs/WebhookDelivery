package subscriptions

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"
)

func (s *Service) Add(ctx context.Context, command dto.AddSubscriptionCommand) (*dto.AddSubscriptionResult, error) {
	const fn = "services.subscriptions.Service.Add"
	log := s.log.With(slog.String("fn", fn))

	subs, err := s.repo.Add(ctx, command)
	if err != nil {
		const msg = "failed to add subscription"
		if utils.IsDomainError(err) || utils.IsCtxError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("endpoint_id", command.EndpointID))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("endpoint_id", command.EndpointID))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("subscriptions were added successfully", slog.String("endpoint_id", command.EndpointID), slog.Int("count", len(command.EventTypes)))

	return &dto.AddSubscriptionResult{
		Subscriptions: subs,
	}, nil
}
