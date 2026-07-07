package app

import (
	"webhook-delivery/internal/app/http"
	"webhook-delivery/internal/config"
)

type App struct {
	Http *http.App
}

func NewApp(cfg config.Config) *App {
	return &App{
		Http: http.New(cfg.HttpConfig),
	}
}
