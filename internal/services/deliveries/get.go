package deliveries

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/services/utils"

	"github.com/google/uuid"
)

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
	const fn = "services.deliveries.Service.GetByID"
	log := s.log.With(slog.String("fn", fn))

	delivery, err := s.repo.GetByID(ctx, id)
	if err != nil {
		const msg = "failed to get delivery by id"
		if utils.IsCtxError(err) || utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("id", id.String()))
			return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("id", id.String()))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("delivery info was sent successfully", slog.String("id", id.String()))

	return delivery, nil
}

func (s *Service) GetFromEvent(ctx context.Context, command dto.GetDeliveriesFromEventCommand) ([]domain.Delivery, int, error) {
	const fn = "services.deliveries.Service.GetFromEvent"
	log := s.log.With(slog.String("fn", fn))

	deliveries, total, err := s.repo.GetFromEvent(ctx, command)
	if err != nil {
		const msg = "failed to get deliveries from event"
		if utils.IsCtxError(err) || utils.IsDomainError(err) {
			log.Info(msg, sloglib.Error(err), slog.String("event_id", command.EventID.String()))
			return nil, 0, fmt.Errorf("%s: %s: %w", fn, msg, err)
		}

		log.Error(msg, sloglib.Error(err), slog.String("event_id", command.EventID.String()))
		return nil, 0, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("deliveries info from event was sent successfully", slog.String("event_id", command.EventID.String()),
		slog.Int("total", total),
		slog.Int("limit", command.Limit),
		slog.Int("page", command.Page))

	return deliveries, total, nil
}
