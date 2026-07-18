package endpoints

import (
	"log/slog"
	"webhook-delivery/internal/infrastructure/repo"
)

type Repo struct {
	db  repo.DB
	log *slog.Logger
}

func NewRepo(db repo.DB, log *slog.Logger) *Repo {
	return &Repo{db: db, log: log}
}
