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
	"webhook-delivery/internal/services/deliveries"
	"webhook-delivery/internal/services/endpoints"
	"webhook-delivery/internal/services/events"
	"webhook-delivery/internal/services/subscriptions"
	"webhook-delivery/internal/services/workers"
)

type App struct {
	Server   *server.Server
	Postgres *postgres.App

	workers    *workers.Worker
	dispatcher *workers.Dispatcher

	log *slog.Logger
	cfg config.Config

	ctx    context.Context
	cancel context.CancelFunc
}

func New(log *slog.Logger, cfg config.Config) *App {
	postgresApp, err := postgres.New(log, cfg.DB)
	if err != nil {
		os.Exit(1)
	}

	postgresApp.MustConnect()
	postgres.MustMigrate(log, cfg.DB)

	eventRepo := pgRepo.NewEvents(postgresApp.Pool, log)
	endpointRepo := pgRepo.NewEndpoints(postgresApp.Pool, log)
	deliveryRepo := pgRepo.NewDeliveries(postgresApp.Pool, log, cfg.MaxAttempts)
	subscriptionRepo := pgRepo.NewSubscriptions(postgresApp.Pool, log)

	worker := workers.NewWorkers(deliveryRepo, cfg.Workers, log)
	dispatcher := workers.NewDispatcher(worker, deliveryRepo, cfg.Dispatcher, log)

	handler := rest.NewHandler(rest.Services{
		Deliveries:    deliveries.NewService(deliveryRepo, log),
		Endpoints:     endpoints.NewService(endpointRepo, log),
		Subscriptions: subscriptions.NewService(subscriptionRepo, log),
		Events:        events.NewService(eventRepo, deliveryRepo, log),
	})

	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		Server:     server.New(handler, cfg, log),
		Postgres:   postgresApp,
		workers:    worker,
		dispatcher: dispatcher,
		log:        log,
		cfg:        cfg,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (a *App) MustRun() {
	const fn = "app.App.MustRun"
	log := a.log.With(slog.String("fn", fn))

	if err := a.dispatcher.Recover(a.ctx); err != nil {
		log.Error("failed to recover deliveries", sloglib.Error(err))
		os.Exit(1)
	}

	go a.workers.Run(a.ctx)
	go a.dispatcher.Run(a.ctx)

	log.Info("workers and dispatcher is running", slog.Int("workers_count", a.cfg.MaxGoroutines))

	a.log.Info("http server is running")
	if err := a.Server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.log.Error("http server failed", sloglib.Error(err))
		os.Exit(1)
	}
}

func (a *App) Shutdown(ctx context.Context) {
	a.cancel()

	if err := a.Server.Shutdown(ctx); err != nil {
		a.log.Error("failed to shutdown http server", sloglib.Error(err))
	}

	a.Postgres.Close()
}
