package app

import (
	"log/slog"
	"os"
	"webhook-delivery/internal/app/http"
	"webhook-delivery/internal/app/pg"
	"webhook-delivery/internal/app/workers"
	"webhook-delivery/internal/config"
	"webhook-delivery/internal/delivery/handlers/endpoints"
	"webhook-delivery/internal/delivery/handlers/events"
	"webhook-delivery/internal/delivery/handlers/subscriptions"
	"webhook-delivery/internal/delivery/middlewares"
	pgRepo "webhook-delivery/internal/infrastructure/pg"
	endpointsSvc "webhook-delivery/internal/services/endpoints"
	eventSvc "webhook-delivery/internal/services/events"
	subscriptionsSvc "webhook-delivery/internal/services/subscriptions"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type App struct {
	HTTP     *http.App
	Postgres *pg.App
	Workers  *workers.App

	cfg config.Config
	log *slog.Logger
}

func New(log *slog.Logger, cfg config.Config) *App {
	postgresApp, err := pg.New(log, cfg.DbConfig)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	pg.MustMigrate(log, cfg.DbConfig)

	deliveryRepo := pgRepo.NewDeliveries(postgresApp.Pool, log, cfg.MaxAttempts)

	return &App{
		HTTP:     http.New(log, cfg.HttpConfig),
		Postgres: postgresApp,
		Workers:  workers.NewApp(deliveryRepo, log, cfg.WorkersConfig),
		cfg:      cfg,
		log:      log,
	}
}

func (a *App) MountMiddlewares() {
	a.HTTP.Router.Use(middleware.RequestID)
	a.HTTP.Router.Use(middleware.RealIP)
	a.HTTP.Router.Use(middleware.Recoverer)
	a.HTTP.Router.Use(middleware.URLFormat)
	a.HTTP.Router.Use(middlewares.NewLogger(a.log))
}

func (a *App) MountHandlers() {
	eventRepo := pgRepo.NewEvents(a.Postgres.Pool, a.log)
	endpointRepo := pgRepo.NewEndpoints(a.Postgres.Pool, a.log)
	deliveryRepo := pgRepo.NewDeliveries(a.Postgres.Pool, a.log, a.cfg.MaxAttempts)
	subscriptionsRepo := pgRepo.NewSubscriptions(a.Postgres.Pool, a.log)

	eventService := eventSvc.NewService(eventRepo, deliveryRepo, a.Workers.Workers, a.log)
	endpointsService := endpointsSvc.NewService(a.log, endpointRepo)
	subscriptionService := subscriptionsSvc.NewService(subscriptionsRepo, a.log)

	a.HTTP.Router.Route("/api/v1", func(r chi.Router) {
		r.Route("/endpoints", func(r chi.Router) {
			r.Delete("/{id}", endpoints.Delete(endpointsService, a.log))
			r.With(middlewares.NewUrlPaginationParser(a.log)).
				Get("/", endpoints.GetAll(endpointsService, a.log))

			r.Get("/{id}", endpoints.Get(endpointsService, a.log))

			r.With(middlewares.NewBodyParser[endpoints.RegisterRequest](a.log)).
				With(middlewares.NewValidator[endpoints.RegisterRequest](a.log)).
				Post("/", endpoints.Register(endpointsService, a.log))

			r.With(middlewares.NewBodyParser[endpoints.UpdateRequest](a.log)).
				With(middlewares.NewValidator[endpoints.UpdateRequest](a.log)).
				Patch("/{id}", endpoints.Update(endpointsService, a.log))

			r.Route("/{id}/subscriptions", func(r chi.Router) {
				r.With(middlewares.NewBodyParser[subscriptions.SubscribeRequest](a.log)).
					With(middlewares.NewValidator[subscriptions.SubscribeRequest](a.log)).
					Post("/", subscriptions.Subscribe(subscriptionService, a.log))

				r.Get("/", subscriptions.GetAll(subscriptionService, a.log))
			})
		})

		r.Route("/events", func(r chi.Router) {
			r.Get("/{id}", events.Get(eventService, a.log))

			r.With(middlewares.NewBodyParser[events.PublishRequest](a.log)).
				With(middlewares.NewValidator[events.PublishRequest](a.log)).
				Post("/", events.Publish(eventService, a.log))
		})

		r.Route("/subscriptions", func(r chi.Router) {
			r.Delete("/{id}", subscriptions.Delete(subscriptionService, a.log))
		})
	})
}
