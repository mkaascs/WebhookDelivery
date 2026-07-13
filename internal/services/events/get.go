package events

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"
)

func (s *Service) Get(ctx context.Context, eventID string) (*dto.GetEventResult, error) {
	const fn = "services.events.Service.Get"
	log := s.log.With(slog.String("fn", fn))

	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		const msg = "failed to get event by id"
		if utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("event_id", eventID))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("event_id", eventID))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	return &dto.GetEventResult{
		Event: *event,
	}, nil
}
