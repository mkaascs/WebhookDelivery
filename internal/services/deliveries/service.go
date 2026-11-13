package deliveries

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/google/uuid"
)

type DeliveryRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error)
	GetFromEvent(ctx context.Context, command dto.GetDeliveriesFromEventCommand) ([]domain.Delivery, int, error)
	UpdateStatus(ctx context.Context, command dto.UpdateDeliveryStatusCommand) error
}

type Service struct {
	log  *slog.Logger
	repo DeliveryRepo
}

func NewService(repo DeliveryRepo, log *slog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) Retry(ctx context.Context, id uuid.UUID) error {
	const fn = "services.deliveries.Service.Retry"
	log := s.log.With(slog.String("fn", fn))

	cmd := dto.UpdateDeliveryStatusCommand{
		ID:               id,
		Status:           domain.StatusPending,
		Attempts:         0,
		NextRetryAt:      time.Now(),
		LastError:        nil,
		LastResponseCode: nil,
	}

	if err := s.repo.UpdateStatus(ctx, cmd); err != nil {
		const msg = "failed to retry delivery"
		s.logError(log, msg, err, slog.String("delivery_id", id.String()))
		return fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("delivery was retried", slog.String("delivery_id", id.String()))

	return nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
	const fn = "services.deliveries.Service.GetByID"
	log := s.log.With(slog.String("fn", fn))

	delivery, err := s.repo.GetByID(ctx, id)
	if err != nil {
		const msg = "failed to get delivery by id"
		s.logError(log, msg, err, slog.String("delivery_id", id.String()))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("delivery info was sent successfully", slog.String("delivery_id", id.String()))

	return delivery, nil
}

func (s *Service) GetFromEvent(ctx context.Context, command dto.GetDeliveriesFromEventCommand) ([]domain.Delivery, int, error) {
	const fn = "services.deliveries.Service.GetFromEvent"
	log := s.log.With(slog.String("fn", fn))

	deliveries, total, err := s.repo.GetFromEvent(ctx, command)
	if err != nil {
		const msg = "failed to get deliveries from event"
		s.logError(log, msg, err, slog.String("event_id", command.EventID.String()))
		return nil, 0, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("deliveries info from event was sent successfully", slog.String("event_id", command.EventID.String()),
		slog.Int("total", total),
		slog.Int("limit", command.Limit),
		slog.Int("page", command.Page))

	return deliveries, total, nil
}

func (s *Service) logError(log *slog.Logger, msg string, err error, args ...any) {
	if domain.IsDomainError(err) || domain.IsCtxError(err) {
		log.With(args...).Info(msg, sloglib.Error(err))
		return
	}

	log.With(args...).Error(msg, sloglib.Error(err))
}
