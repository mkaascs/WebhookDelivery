package workers

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
	"webhook-delivery/internal/config"
	"webhook-delivery/internal/domain/dto"
)

type DeliveryRepo interface {
	ClaimPending(ctx context.Context, batchSize int) ([]dto.ClaimPendingResult, error)
	UpdateStatus(ctx context.Context, command dto.UpdateDeliveryStatusCommand) error
}

type Service struct {
	log          *slog.Logger
	cfg          config.WorkersConfig
	deliveryRepo DeliveryRepo
	httpClient   *http.Client

	ctx    context.Context
	cancel context.CancelFunc

	wg     sync.WaitGroup
	notify chan struct{}
	ticker *time.Ticker
}

func NewService(repo DeliveryRepo, log *slog.Logger, cfg config.WorkersConfig) *Service {
	notify := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	httpClient := &http.Client{Timeout: cfg.Timeout}

	return &Service{
		cfg:          cfg,
		log:          log,
		deliveryRepo: repo,
		httpClient:   httpClient,
		notify:       notify,
		ctx:          ctx,
		cancel:       cancel,
		wg:           sync.WaitGroup{},
		ticker:       time.NewTicker(cfg.TickerDuration),
	}
}

func (s *Service) Run() {
	const fn = "services.workers.Service.Run"
	log := s.log.With(slog.String("fn", fn))

	for count := 0; count < s.cfg.MaxGoroutines; count++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()

			for {
				select {
				case <-s.ctx.Done():
					return
				case <-s.notify:
				case <-s.ticker.C:
				}

				s.processBatch(s.ctx)
			}
		}()
	}

	log.Info("request workers is running", slog.Int("count", s.cfg.MaxGoroutines))
}

func (s *Service) Shutdown() {
	const fn = "services.workers.Service.Shutdown"
	log := s.log.With(slog.String("fn", fn))

	s.ticker.Stop()
	s.cancel()
	s.wg.Wait()

	log.Info("request workers was graceful shutdown")
}

func (s *Service) Notify() {
	select {
	case s.notify <- struct{}{}:
	default:
	}
}
