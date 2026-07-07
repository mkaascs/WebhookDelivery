package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"webhook-delivery/internal/config"
)

type App struct {
	mux    *http.ServeMux
	server *http.Server
}

func (a *App) Run() error {
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

func New(cfg config.HttpConfig) *App {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &App{
		mux:    mux,
		server: server,
	}
}
