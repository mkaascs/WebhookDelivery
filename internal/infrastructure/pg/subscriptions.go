package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/lib/uuid"
)

func insertSubscription(ctx context.Context, pool poolQuery, endpointID string, eventTypes []string) error {
	const fn = "infrastructure.pg.insertSubscription"

	batch := &pgx.Batch{}
	for _, eventType := range eventTypes {
		batch.Queue(`
			INSERT INTO subscriptions(id, endpoint_id, event_type) 
			VALUES ($1, $2, $3)`,
			uuid.NewSubscription(), endpointID, eventType)
	}

	res := pool.SendBatch(ctx, batch)
	defer res.Close()

	for range eventTypes {
		if _, err := res.Exec(); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				return domain.ErrSubscriptionAlreadyExists
			}

			return fmt.Errorf("%s: %w", fn, err)
		}
	}

	return nil
}

func getEndpointEventTypes(ctx context.Context, pool poolQuery, endpointID string) ([]string, error) {
	const fn = "infrastructure.pg.getEndpointEventTypes"

	rows, err := pool.Query(ctx, `
		SELECT event_types
		FROM subscriptions
		WHERE endpoint_id = $1`, endpointID)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	defer rows.Close()

	eventTypes := make([]string, 0)
	for rows.Next() {
		var eventType string
		if err := rows.Scan(&eventType); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		eventTypes = append(eventTypes, eventType)
	}

	return eventTypes, nil
}
