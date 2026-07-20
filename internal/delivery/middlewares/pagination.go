package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
)

type PaginationParams struct {
	Page  int
	Limit int
}

const (
	minLimit = 10
	maxLimit = 100
)

type paginationCtxKey int

const paginationParamsKey paginationCtxKey = 0

func NewUrlPaginationParser(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(slog.String("component", "middleware/pagination"))

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			page, err := strconv.Atoi(req.URL.Query().Get("page"))
			if err != nil || page < 1 {
				page = 1
			}

			limit, err := strconv.Atoi(req.URL.Query().Get("limit"))
			if err != nil || limit < 1 {
				limit = minLimit
			} else if limit > maxLimit {
				limit = maxLimit
			}

			params := &PaginationParams{
				Page:  page,
				Limit: limit,
			}

			ctx := context.WithValue(req.Context(), paginationParamsKey, params)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

func GetPaginationParams(context context.Context) (*PaginationParams, bool) {
	params, ok := context.Value(paginationParamsKey).(*PaginationParams)
	return params, ok
}
