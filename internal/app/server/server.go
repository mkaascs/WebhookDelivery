package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/config"
	"webhook-delivery/internal/delivery/rest/middlewares"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/swaggest/swgui/v5emb"
)

type Server struct {
	http *http.Server
}

func New(handlers apiv1.ServerInterface, cfg config.Config, log *slog.Logger) *Server {
	spec, err := apiv1.GetSpec()
	if err != nil {
		log.Error("failed to get spec", sloglib.Error(err))
		os.Exit(1)
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middlewares.NewLogger(log))

	if cfg.Env != config.EnvProd {
		router.Get("/swagger/openapi.json", newSpecHandler(spec))
		router.Mount("/swagger", v5emb.New(
			"Webhook Delivery",
			"/swagger/openapi.json",
			"/swagger/"))
	}

	validator := nethttpmiddleware.OapiRequestValidatorWithOptions(spec,
		&nethttpmiddleware.Options{
			SilenceServersWarning: true,
			ErrorHandler:          validationErrorHandler,
			Options: openapi3filter.Options{
				AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			},
		})

	// group keeps the validator off the swagger routes and lets it be added after them
	router.Group(func(r chi.Router) {
		r.Use(validator)
		apiv1.HandlerFromMuxWithBaseURL(handlers, r, "/api/v1")
	})

	return &Server{http: &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  cfg.ReadTimeout},
	}
}

func (s *Server) Run() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func newSpecHandler(spec *openapi3.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(spec)
	}
}

func validationErrorHandler(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var code string
	switch statusCode {
	case http.StatusBadRequest:
		code = "bad_request"
		message = cleanValidationMessage(message)
	case http.StatusNotFound:
		code = "not_found"
		message = "resource not found"
	default:
		code = "error"
	}

	_ = json.NewEncoder(w).Encode(apiv1.Error{
		Code:    &code,
		Message: message,
	})
}

func cleanValidationMessage(raw string) string {
	if _, after, ok := strings.Cut(raw, `Error at "/`); ok {
		rest := after
		if before, _, ok := strings.Cut(rest, `"`); ok {
			return "invalid field: " + before
		}
	}

	return "invalid request"
}
