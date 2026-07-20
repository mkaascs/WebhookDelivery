package workers

import (
	"log/slog"
	"webhook-delivery/internal/config"
	pgRepo "webhook-delivery/internal/infrastructure/pg"
	"webhook-delivery/internal/services/workers"
)

type App struct {
	Workers *workers.Service
}

func NewApp(repo *pgRepo.Deliveries, log *slog.Logger, cfg config.WorkersConfig) *App {
	return &App{
		Workers: workers.NewService(repo, log, cfg),
	}
}

func (a *App) Start() {
	a.Workers.Run()
}

func (a *App) Shutdown() {
	a.Workers.Shutdown()
}
