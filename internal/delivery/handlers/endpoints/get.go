package endpoints

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/lib/ptr"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type EndpointInfo struct {
	ID          string    `json:"id" example:"ep_1a2b3c"`
	URL         string    `json:"url" example:"https://example.com/webhooks"`
	EventTypes  []string  `json:"event_types" example:"order.created"`
	Description string    `json:"description" example:"Billing webhook"`
	IsActive    bool      `json:"is_active" example:"true"`
	CreatedAt   time.Time `json:"created_at" example:"2026-07-01T12:00:00Z"`
}

type GetResponse struct {
	utils.Response
	EndpointInfo
}

type EndpointGetter interface {
	GetByID(ctx context.Context, id string) (*dto.GetEndpointResult, error)
}

// Get godoc
//
//	@Summary	Get an endpoint by ID
//	@Tags		endpoints
//	@Produce	json
//	@Param		id	path		string	true	"Endpoint ID"
//	@Success	200	{object}	GetResponse
//	@Failure	400	{object}	utils.Response	"id is not provided"
//	@Failure	404	{object}	utils.Response	"endpoint not found"
//	@Failure	500	{object}	utils.Response
//	@Router		/endpoints/{id} [get]
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
			if utils.IsCtxError(err) || utils.TryRenderDomainError(w, req, err) {
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
				EventTypes:  utils.GetDefaultIfNull(result.EventTypes),
				Description: ptr.Defer(result.Description),
				IsActive:    result.IsActive,
				CreatedAt:   result.CreatedAt,
			},
		})
	}
}
