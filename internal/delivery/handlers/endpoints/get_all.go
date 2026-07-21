package endpoints

import (
	"context"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/lib/ptr"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type GetAllResponse struct {
	utils.Response
	Page      int            `json:"page" example:"1"`   // Current page (from 1)
	Limit     int            `json:"limit" example:"10"` // Page size
	Total     int            `json:"total" example:"42"` // Total number of endpoints
	Endpoints []EndpointInfo `json:"endpoints"`
}

type AllEndpointsGetter interface {
	GetAll(ctx context.Context, command dto.GetAllEndpointsCommand) (*dto.GetAllEndpointsResult, error)
}

// GetAll godoc
//
//	@Summary		List endpoints
//	@Description	Returns a paginated list of endpoints. Invalid or missing pagination params fall back to defaults.
//	@Tags			endpoints
//	@Produce		json
//	@Param			page	query		int	false	"Page number (from 1, default 1)"
//	@Param			limit	query		int	false	"Page size (10-100, default 10)"
//	@Success		200		{object}	GetAllResponse
//	@Failure		500		{object}	utils.Response
//	@Router			/endpoints [get]
func GetAll(getter AllEndpointsGetter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.endpoints.GetAll"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		pagination, ok := middlewares.GetPaginationParams(req.Context())
		if !ok {
			log.Error("failed to get pagination params. make sure that PaginationParserMiddleware is enabled")
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		result, err := getter.GetAll(req.Context(), dto.GetAllEndpointsCommand{
			Page:  pagination.Page,
			Limit: pagination.Limit,
		})

		if err != nil {
			const msg = "failed to get all endpoints"
			if utils.IsCtxError(err) || utils.TryRenderDomainError(w, req, err) {
				log.Info(msg, sloglib.Error(err))
				return
			}

			log.Error(msg, sloglib.Error(err))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		endpointsInfo := make([]EndpointInfo, 0, len(result.Endpoints))
		for _, info := range result.Endpoints {
			endpointsInfo = append(endpointsInfo, EndpointInfo{
				ID:          info.ID,
				URL:         info.URL,
				EventTypes:  utils.GetDefaultIfNull(info.EventTypes),
				Description: ptr.Defer(info.Description),
				IsActive:    info.IsActive,
				CreatedAt:   info.CreatedAt,
			})
		}

		render.Status(req, http.StatusOK)
		render.JSON(w, req, GetAllResponse{
			Page:      pagination.Page,
			Limit:     pagination.Limit,
			Total:     result.Total,
			Endpoints: endpointsInfo,
		})
	}
}
