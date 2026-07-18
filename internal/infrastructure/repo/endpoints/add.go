package endpoints

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
	"webhook-delivery/internal/lib/uuid"
)

func (r *Repo) Add(ctx context.Context, command dto.AddEndpointCommand) (*dto.AddEndpointResult, error) {
	const fn = "repo.endpoints.Repo.Add"
	log := r.log.With(slog.String("fn", fn))

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	success := false
	defer func(tx *sql.Tx) {
		if !success {
			if err := tx.Rollback(); err != nil {
				log.Error("failed to rollback tx", sloglib.Error(err))
			}
		}
	}(tx)

	var endpointID string
	var createdAt time.Time

	err = tx.QueryRowContext(ctx, `
		INSERT INTO endpoints (id, url, secret, description, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`).
		Scan(&endpointID, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO subscriptions (id, endpoint_id, event_type)
		VALUES ($1, $2, $3)`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			log.Warn("failed to close statement", sloglib.Error(err))
		}
	}(stmt)

	for _, eventType := range command.EventTypes {
		_, err := stmt.ExecContext(ctx, uuid.NewSubscription(), endpointID, eventType)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
	}

	success = true
	return &dto.AddEndpointResult{
		ID:        endpointID,
		CreatedAt: createdAt,
	}, nil
}
