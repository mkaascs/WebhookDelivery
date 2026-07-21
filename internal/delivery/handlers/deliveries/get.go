package deliveries

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type DeliveryInfo struct {
	ID               string    `json:"id"`
	EndpointID       string    `json:"endpoint_id"`
	EventID          string    `json:"event_id"`
	Status           string    `json:"status"`
	Attempts         int       `json:"attempts"`
	MaxAttempts      int       `json:"max_attempts"`
	NextRetryAt      time.Time `json:"next_retry_at"`
	CreatedAt        time.Time `json:"created_at"`
	LastResponseCode *int      `json:"last_response_code"`
	LastError        *string   `json:"last_error"`
}

type GetResponse struct {
	DeliveryInfo `json:"delivery"`
}

type DeliveryGetter interface {
	GetByID(ctx context.Context, id string) (*dto.GetDeliveryResult, error)
}

func Get(getter DeliveryGetter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.deliveries.Get"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		id := chi.URLParam(req, "id")
		if id == "" {
			log.Info("no delivery id provided")
			utils.RenderError(w, req, http.StatusBadRequest, "url param delivery id is not provided")
			return
		}

		result, err := getter.GetByID(req.Context(), id)
		if err != nil {
			const msg = "failed to get delivery by id"
			if utils.IsCtxError(err) || utils.TryRenderDomainError(w, req, err) {
				log.Info(msg, sloglib.Error(err), slog.String("id", id))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("id", id))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(w, req, GetResponse{
			DeliveryInfo: mapDomainToResponse(result.Delivery),
		})
	}
}

func mapDomainToResponse(result domain.Delivery) DeliveryInfo {
	return DeliveryInfo{
		ID:               result.ID,
		EndpointID:       result.EndpointID,
		EventID:          result.EventID,
		Status:           string(result.Status),
		Attempts:         result.Attempts,
		MaxAttempts:      result.MaxAttempts,
		NextRetryAt:      result.NextRetryAt,
		CreatedAt:        result.CreatedAt,
		LastResponseCode: result.LastResponseCode,
		LastError:        result.LastError,
	}
}
