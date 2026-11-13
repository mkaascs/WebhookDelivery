package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"webhook-delivery/internal/config"
	"webhook-delivery/internal/domain"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type DispatcherRepo interface {
	ClaimPending(ctx context.Context, limit int) ([]domain.PendingDelivery, error)
	RecoverProcessing(ctx context.Context) error
}

type Queue interface {
	Enqueue(delivery domain.PendingDelivery)
	FreeSlots() int
}

type Dispatcher struct {
	queue Queue
	repo  DispatcherRepo

	cfg config.Dispatcher
	log *slog.Logger
}

func NewDispatcher(queue Queue, repo DispatcherRepo, cfg config.Dispatcher, log *slog.Logger) *Dispatcher {
	return &Dispatcher{
		queue: queue,
		repo:  repo,
		cfg:   cfg,
		log:   log,
	}
}

func (d *Dispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(d.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.dispatch(ctx)
		}
	}
}

func (d *Dispatcher) Recover(ctx context.Context) error {
	const fn = "services.workers.Dispatcher.Recover"
	log := d.log.With(slog.String("fn", fn))

	err := d.repo.RecoverProcessing(ctx)
	if err != nil {
		log.Error("failed to recover processing", sloglib.Error(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (d *Dispatcher) dispatch(ctx context.Context) {
	const fn = "services.workers.Dispatcher.dispatch"
	log := d.log.With(slog.String("fn", fn))

	free := min(d.queue.FreeSlots(), d.cfg.BatchSize)
	if free == 0 {
		return
	}

	deliveries, err := d.repo.ClaimPending(ctx, free)
	if err != nil {
		log.Error("failed to claim pending deliveries", sloglib.Error(err))
		return
	}

	for _, delivery := range deliveries {
		d.queue.Enqueue(delivery)
	}

	count := len(deliveries)
	if count > 0 {
		log.Info("dispatched pending deliveries", slog.Int("count", count))
	}
}
