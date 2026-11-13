package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"webhook-delivery/internal/app/postgres"
	"webhook-delivery/internal/app/server"
	"webhook-delivery/internal/config"
	"webhook-delivery/internal/delivery/rest"
	pgRepo "webhook-delivery/internal/infrastructure/pg"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	deliverySvc "webhook-delivery/internal/services/deliveries"
	endpointSvc "webhook-delivery/internal/services/endpoints"
	eventSvc "webhook-delivery/internal/services/events"
	subscriptionSvc "webhook-delivery/internal/services/subscriptions"
	"webhook-delivery/internal/services/workers"
)

type App struct {
	Server   *server.Server
	Postgres *postgres.App
	Workers  *workers.Service

	log *slog.Logger
}

func New(log *slog.Logger, cfg config.Config) *App {
	postgresApp, err := postgres.New(log, cfg.DbConfig)
	if err != nil {
		os.Exit(1)
	}

	postgresApp.MustConnect()
	postgres.MustMigrate(log, cfg.DbConfig)

	eventRepo := pgRepo.NewEvents(postgresApp.Pool, log)
	endpointRepo := pgRepo.NewEndpoints(postgresApp.Pool, log)
	deliveryRepo := pgRepo.NewDeliveries(postgresApp.Pool, log, cfg.MaxAttempts)
	subscriptionRepo := pgRepo.NewSubscriptions(postgresApp.Pool, log)

	workersService := workers.NewService(deliveryRepo, log, cfg.WorkersConfig)

	handler := rest.NewHandler(rest.Services{
		Deliveries:    deliverySvc.NewService(deliveryRepo, workersService, log),
		Endpoints:     endpointSvc.NewService(endpointRepo, log),
		Subscriptions: subscriptionSvc.NewService(subscriptionRepo, log),
		Events:        eventSvc.NewService(eventRepo, deliveryRepo, workersService, log),
	})

	return &App{
		Server:   server.New(handler, cfg, log),
		Postgres: postgresApp,
		Workers:  workersService,
		log:      log,
	}
}

func (a *App) MustRun() {
	go a.Workers.Run()

	a.log.Info("http server is running")
	if err := a.Server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.log.Error("http server failed", sloglib.Error(err))
		os.Exit(1)
	}
}

func (a *App) Shutdown(ctx context.Context) {
	if err := a.Server.Shutdown(ctx); err != nil {
		a.log.Error("failed to shutdown http server", sloglib.Error(err))
	}

	a.Workers.Shutdown()
	a.Postgres.Close()
}
