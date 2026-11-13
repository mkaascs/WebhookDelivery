package workers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"webhook-delivery/internal/config"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type DeliveryRepo interface {
	UpdateStatus(ctx context.Context, command dto.UpdateDeliveryStatusCommand) error
}

type Worker struct {
	httpClient   *http.Client
	deliveryRepo DeliveryRepo

	log *slog.Logger
	cfg config.Workers

	wg    sync.WaitGroup
	queue chan *domain.PendingDelivery
}

func NewWorkers(repo DeliveryRepo, cfg config.Workers, log *slog.Logger) *Worker {
	return &Worker{
		httpClient:   &http.Client{Timeout: cfg.Timeout},
		deliveryRepo: repo,
		log:          log,
		cfg:          cfg,
		wg:           sync.WaitGroup{},
		queue:        make(chan *domain.PendingDelivery, cfg.QueueSize),
	}
}

func (w *Worker) Run(ctx context.Context) {
	for count := 0; count < w.cfg.MaxGoroutines; count++ {
		w.wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case delivery := <-w.queue:
					w.process(ctx, delivery)
				}
			}
		})
	}

	w.wg.Wait()
}

func (w *Worker) Enqueue(delivery *domain.PendingDelivery) {
	w.queue <- delivery
}

func (w *Worker) FreeSlots() int {
	return cap(w.queue) - len(w.queue)
}

func (w *Worker) process(ctx context.Context, delivery *domain.PendingDelivery) {
	const fn = "services.workers.Worker.process"
	log := w.log.With(slog.String("fn", fn))

	delivery.Attempts++

	code, msg, err := w.sendPostRequest(ctx, delivery.URL, delivery.Payload, []byte(delivery.Secret))
	if err != nil {
		log.Warn("failed to send post request", slog.String("delivery_id", delivery.ID.String()), sloglib.Error(err))
	}

	if isSuccessCode(code) {
		log.Info("successfully processed delivery", slog.String("delivery_id", delivery.ID.String()))
		err := w.deliveryRepo.UpdateStatus(ctx, dto.UpdateDeliveryStatusCommand{
			ID:          delivery.ID,
			Status:      domain.StatusDelivered,
			Attempts:    delivery.Attempts,
			NextRetryAt: delivery.NextRetryAt,
		})

		if err != nil {
			log.Warn("failed to update delivery", slog.String("delivery_id", delivery.ID.String()), sloglib.Error(err))
		}

		return
	}

	log.Info("failed to process delivery", slog.String("delivery_id", delivery.ID.String()), slog.Int("status_code", code), slog.Int("attempt", delivery.Attempts))

	command := dto.UpdateDeliveryStatusCommand{
		ID:               delivery.ID,
		Status:           domain.StatusFailed,
		Attempts:         delivery.Attempts,
		NextRetryAt:      delivery.NextRetryAt,
		LastError:        &msg,
		LastResponseCode: &code,
	}

	if delivery.Attempts < w.cfg.MaxAttempts {
		command.Status = domain.StatusPending
		command.NextRetryAt = time.Now().Add(w.backoff(delivery.Attempts))
	}

	if err := w.deliveryRepo.UpdateStatus(ctx, command); err != nil {
		log.Warn("failed to update delivery", slog.String("delivery_id", delivery.ID.String()), sloglib.Error(err))
	}
}

func (w *Worker) backoff(attempts int) time.Duration {
	multiplier := 1 << (attempts - 1)
	delay := float64(w.cfg.BaseBackoff) * float64(multiplier)
	if delay > float64(w.cfg.MaxBackoff) {
		delay = float64(w.cfg.MaxBackoff)
	}

	jitter := delay * 0.2 * (rand.Float64()*2 - 1)
	delay += jitter

	return time.Duration(delay)
}

func (w *Worker) sendPostRequest(ctx context.Context, url string, payload json.RawMessage, secret []byte) (_ int, _ string, err error) {
	hash := hmac.New(sha256.New, secret)
	hash.Write(payload)
	signature := hex.EncodeToString(hash.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, err.Error(), err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return 0, err.Error(), err
	}

	defer func(body io.ReadCloser) {
		if err == nil {
			err = body.Close()
			return
		}

		_ = body.Close()
	}(resp.Body)

	return resp.StatusCode, fmt.Sprintf("HTTP %d", resp.StatusCode), nil
}

func isSuccessCode(httpCode int) bool {
	return httpCode >= 200 && httpCode < 400
}
