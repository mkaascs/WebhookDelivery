package endpoints

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/lib/ptr"
)

type EndpointInfo struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	EventTypes  []string  `json:"event_types"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type GetResponse struct {
	utils.Response
	EndpointInfo
}

type EndpointGetter interface {
	GetByID(ctx context.Context, id string) (*dto.GetEndpointResult, error)
}

func Get(getter EndpointGetter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.endpoints.Get"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		id := chi.URLParam(req, "id")
		if id == "" {
			log.Info("no endpoint id provided")
			utils.RenderError(w, req, http.StatusBadRequest, "url param endpoint id is not provided")
			return
		}

		result, err := getter.GetByID(req.Context(), id)
		if err != nil {
			const msg = "failed to get endpoint by id"
			if utils.IsCtxError(err) || utils.TryRenderEndpointsError(w, req, err) {
				log.Info(msg, sloglib.Error(err))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("endpoint_id", id))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(w, req, GetResponse{
			EndpointInfo: EndpointInfo{
				ID:          id,
				URL:         result.URL,
				EventTypes:  result.EventTypes,
				Description: ptr.Defer(result.Description),
				IsActive:    result.IsActive,
				CreatedAt:   result.CreatedAt,
			},
		})
	}
}
