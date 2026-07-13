package events

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"
)

func (s *Service) Publish(ctx context.Context, command dto.PublishEventCommand) (*dto.PublishEventResult, error) {
	const fn = "services.events.Service.Publish"
	log := s.log.With(slog.String("fn", fn))

	event, err := s.eventRepo.Add(ctx, command)
	if err != nil {
		log.Error("failed to add event", sloglib.Error(err))
		return nil, fmt.Errorf("%s: failed to add event: %w", fn, err)
	}

	count, err := s.deliveryRepo.CreateForEvent(ctx, event.ID, event.Type)
	if err != nil {
		const msg = "failed to create deliveries for event"
		if utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("event_id", event.ID))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("event_id", event.ID))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	s.notifier.Notify()

	return &dto.PublishEventResult{
		Event:             *event,
		DeliveriesCreated: count,
	}, nil
}
