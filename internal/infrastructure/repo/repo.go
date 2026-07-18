package repo

import (
	"context"
	"database/sql"
)

type DB interface {
	Begin(ctx context.Context) (*sql.Tx, error)
}
