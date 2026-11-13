package events

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"

	"github.com/google/uuid"
)

func (s *Service) Get(ctx context.Context, eventID uuid.UUID) (*domain.Event, error) {
	const fn = "services.events.Service.Get"
	log := s.log.With(slog.String("fn", fn))

	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		const msg = "failed to get event by id"
		if utils.IsDomainError(err) || utils.IsCtxError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("id", eventID.String()))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("id", eventID.String()))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("event info was sent successfully", slog.String("id", event.ID.String()))

	return event, nil
}
