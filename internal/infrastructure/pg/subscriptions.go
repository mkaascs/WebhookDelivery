package pg

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	"webhook-delivery/internal/lib/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Subscriptions struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewSubscriptions(pool *pgxpool.Pool, log *slog.Logger) *Subscriptions {
	return &Subscriptions{pool: pool, log: log}
}

func (s *Subscriptions) Add(ctx context.Context, command dto.AddSubscriptionCommand) ([]domain.Subscription, error) {
	subs, err := insertSubscription(ctx, s.pool, command.EndpointID, command.EventTypes)
	return subs, err
}

func (s *Subscriptions) GetAll(ctx context.Context, endpointID string) ([]domain.Subscription, error) {
	subs, err := getEndpointSubscriptions(ctx, s.pool, endpointID)
	return subs, err
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

func insertSubscription(ctx context.Context, pool poolQuery, endpointID string, eventTypes []string) ([]domain.Subscription, error) {
	const fn = "infrastructure.pg.insertSubscription"

	subIDs := make([]string, 0, len(eventTypes))
	for range eventTypes {
		subIDs = append(subIDs, uuid.NewSubscription())
	}

	rows, err := pool.Query(ctx, `
		INSERT INTO subscriptions(id, endpoint_id, event_type)
		SELECT unnest($1::text[]), $2, unnest($3::text[])
		ON CONFLICT (endpoint_id, event_type) DO NOTHING
		RETURNING id, event_type, created_at`,
		subIDs, endpointID, eventTypes)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	defer rows.Close()

	subs := make([]domain.Subscription, 0)
	for rows.Next() {
		sub := domain.Subscription{EndpointID: endpointID}
		if err := rows.Scan(&sub.ID, &sub.EventType, &sub.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		subs = append(subs, sub)
	}

	return subs, nil
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
