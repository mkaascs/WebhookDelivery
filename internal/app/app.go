package app

import (
	"log/slog"
	"os"
	"webhook-delivery/internal/app/http"
	"webhook-delivery/internal/app/pg"
	"webhook-delivery/internal/config"
)

type App struct {
	Http     *http.App
	Postgres *pg.App
}

func New(log *slog.Logger, cfg config.Config) *App {
	postgresApp, err := pg.New(log, cfg.DbConfig)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	pg.MustMigrate(log, cfg.DbConfig)

	return &App{
		Http:     http.New(log, cfg.HttpConfig),
		Postgres: postgresApp,
	}
}
