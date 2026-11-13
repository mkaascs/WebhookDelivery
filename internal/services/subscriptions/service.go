package subscriptions

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/google/uuid"
)

type SubscriptionRepo interface {
	Add(ctx context.Context, command dto.AddSubscriptionCommand) ([]domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context, endpointID uuid.UUID) ([]domain.Subscription, error)
}

type Service struct {
	log  *slog.Logger
	repo SubscriptionRepo
}

func NewService(repo SubscriptionRepo, log *slog.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) GetAll(ctx context.Context, endpointID uuid.UUID) ([]domain.Subscription, error) {
	const fn = "services.subscriptions.Service.GetAll"
	log := s.log.With(slog.String("fn", fn))

	subs, err := s.repo.GetAll(ctx, endpointID)
	if err != nil {
		const msg = "failed to get all subscriptions of endpoint"
		s.logError(log, msg, err, slog.String("endpoint_id", endpointID.String()))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("subscriptions info were sent successfully", slog.String("endpoint_id", endpointID.String()))
	return subs, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	const fn = "services.subscriptions.Service.Delete"
	log := s.log.With(slog.String("fn", fn))

	if err := s.repo.Delete(ctx, id); err != nil {
		const msg = "failed to delete subscription"
		s.logError(log, msg, err, slog.String("subscription_id", id.String()))
		return fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("subscription was deleted successfully", slog.String("subscription_id", id.String()))

	return nil
}

func (s *Service) Add(ctx context.Context, command dto.AddSubscriptionCommand) ([]domain.Subscription, error) {
	const fn = "services.subscriptions.Service.Add"
	log := s.log.With(slog.String("fn", fn))

	subs, err := s.repo.Add(ctx, command)
	if err != nil {
		const msg = "failed to add subscription"
		s.logError(log, msg, err, slog.String("endpoint_id", command.EndpointID.String()))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	log.Info("subscriptions were added successfully", slog.String("endpoint_id", command.EndpointID.String()), slog.Int("count", len(command.EventTypes)))

	return subs, nil
}

func (s *Service) logError(log *slog.Logger, msg string, err error, args ...any) {
	if domain.IsDomainError(err) || domain.IsCtxError(err) {
		log.With(args...).Info(msg, sloglib.Error(err))
		return
	}

	log.With(args...).Error(msg, sloglib.Error(err))
}
