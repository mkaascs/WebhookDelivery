package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	"webhook-delivery/internal/lib/uuid"
)

type Subscriptions struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewSubscriptions(pool *pgxpool.Pool, log *slog.Logger) *Subscriptions {
	return &Subscriptions{pool: pool, log: log}
}

func (s *Subscriptions) Add(ctx context.Context, command dto.AddSubscriptionCommand) ([]domain.Subscription, error) {
	if err := insertSubscription(ctx, s.pool, command.EndpointID, command.EventTypes); err != nil {
		return nil, err
	}

	result, err := getEndpointSubscriptions(ctx, s.pool, command.EndpointID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Subscriptions) Delete(ctx context.Context, id string) error {
	const fn = "infrastructure.pg.Subscriptions.Delete"

	res, err := s.pool.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if res.RowsAffected() == 0 {
		return domain.ErrSubscriptionNotFound
	}

	return nil
}

func insertSubscription(ctx context.Context, pool poolQuery, endpointID string, eventTypes []string) (err error) {
	const fn = "infrastructure.pg.insertSubscription"

	batch := &pgx.Batch{}
	for _, eventType := range eventTypes {
		batch.Queue(`
			INSERT INTO subscriptions(id, endpoint_id, event_type) 
			VALUES ($1, $2, $3)`,
			uuid.NewSubscription(), endpointID, eventType)
	}

	res := pool.SendBatch(ctx, batch)
	defer func(res pgx.BatchResults) {
		if err != nil {
			_ = res.Close()
			return
		}

		err = res.Close()
	}(res)

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

func getEndpointSubscriptions(ctx context.Context, pool poolQuery, endpointID string) ([]domain.Subscription, error) {
	const fn = "infrastructure.pg.getEndpointSubscriptions"

	rows, err := pool.Query(ctx, `
		SELECT id, event_type, created_at
		FROM subscriptions
		WHERE endpoint_id = $1`, endpointID)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	defer rows.Close()

	subs := make([]domain.Subscription, 0)
	for rows.Next() {
		s := domain.Subscription{EndpointID: endpointID}
		if err := rows.Scan(&s.ID, &s.EventType, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		subs = append(subs, s)
	}

	return subs, nil
}
