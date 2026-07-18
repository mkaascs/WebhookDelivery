package endpoints

import (
	"webhook-delivery/internal/infrastructure/pg"
)

type Repo struct {
	db pg.DB
}

func NewRepo(db pg.DB) *Repo {
	return &Repo{db: db}
}
