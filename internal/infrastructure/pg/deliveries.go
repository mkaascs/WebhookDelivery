package pg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Deliveries struct {
	maxAttempts int
	pool        *pgxpool.Pool
	log         *slog.Logger
}

func NewDeliveries(pool *pgxpool.Pool, log *slog.Logger, maxAttempts int) *Deliveries {
	return &Deliveries{pool: pool, log: log, maxAttempts: maxAttempts}
}

func (d *Deliveries) CreateForEvent(ctx context.Context, eventID uuid.UUID, eventType string) (int, error) {
	const fn = "infrastructure.pg.Deliveries.CreateForEvent"

	rows, err := d.pool.Query(ctx, `
		SELECT endpoint_id
		FROM subscriptions
		JOIN endpoints ON endpoints.id = endpoint_id
		WHERE event_type = $1 AND endpoints.is_active`, eventType)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	defer rows.Close()

	var deliveryIDs, endpointIDs []uuid.UUID
	for rows.Next() {
		var endpointID uuid.UUID
		if err := rows.Scan(&endpointID); err != nil {
			return 0, fmt.Errorf("%s: %w", fn, err)
		}

		endpointIDs = append(endpointIDs, endpointID)
		deliveryIDs = append(deliveryIDs, uuid.Must(uuid.NewV7()))
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	if len(endpointIDs) == 0 {
		return 0, nil
	}

	res, err := d.pool.Exec(ctx, `
		INSERT INTO deliveries (id, endpoint_id, event_id, max_attempts)
		SELECT d_id, e_id, $3, $4
		FROM unnest($1::uuid[], $2::uuid[]) AS t(d_id, e_id)
		ON CONFLICT (endpoint_id, event_id) DO NOTHING`,
		deliveryIDs, endpointIDs, eventID, d.maxAttempts)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return int(res.RowsAffected()), nil
}

func (d *Deliveries) ClaimPending(ctx context.Context, limit int) ([]domain.PendingDelivery, error) {
	const fn = "infrastructure.pg.Deliveries.ClaimPending"

	rows, err := d.pool.Query(ctx, `
		UPDATE deliveries d
		SET status = 'queued'
		FROM events ev, endpoints e
		WHERE ev.id = d.event_id
		    AND e.id = d.endpoint_id
		    AND d.id IN (
		    	SELECT id 
		    	FROM deliveries
		    	WHERE status = 'pending' AND next_retry_at <= now()
		    	ORDER BY next_retry_at
		    	LIMIT $1
		    )
		RETURNING d.id, d.endpoint_id, d.event_id, d.status, d.attempts, d.max_attempts, d.next_retry_at, d.created_at,
		    d.last_response_code, d.last_error, e.url, e.secret, ev.payload`,
		limit)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	defer rows.Close()

	var results []domain.PendingDelivery
	for rows.Next() {
		var result domain.PendingDelivery
		if err := rows.Scan(&result.ID, &result.EndpointID, &result.EventID, &result.Status, &result.Attempts,
			&result.MaxAttempts, &result.NextRetryAt, &result.CreatedAt, &result.LastResponseCode, &result.LastError,
			&result.URL, &result.Secret, &result.Payload); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return results, nil
}

func (d *Deliveries) RecoverProcessing(ctx context.Context) error {
	const fn = "infrastructure.pg.Deliveries.RecoverProcessing"

	_, err := d.pool.Exec(ctx, `UPDATE deliveries SET status = 'pending' WHERE status IN ('queued', 'processing')`)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (d *Deliveries) UpdateStatus(ctx context.Context, command dto.UpdateDeliveryStatusCommand) error {
	const fn = "infrastructure.pg.Deliveries.UpdateStatus"

	res, err := d.pool.Exec(ctx,
		`UPDATE deliveries
			SET status = $1, attempts = $2, next_retry_at = $3, last_error = $5, last_response_code = $6
			WHERE id = $4`,
		command.Status, command.Attempts, command.NextRetryAt, command.ID, command.LastError, command.LastResponseCode)

	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if res.RowsAffected() == 0 {
		return domain.ErrDeliveryNotFound
	}

	return nil
}

func (d *Deliveries) GetByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
	const fn = "infrastructure.pg.Deliveries.GetByID"

	del := &domain.Delivery{ID: id}
	err := d.pool.QueryRow(ctx, `
		SELECT endpoint_id, event_id, status, attempts, max_attempts, next_retry_at, created_at, last_response_code, last_error
		FROM deliveries
		WHERE id = $1`, id).
		Scan(&del.EndpointID, &del.EventID, &del.Status, &del.Attempts, &del.MaxAttempts, &del.NextRetryAt, &del.CreatedAt, &del.LastResponseCode, &del.LastError)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDeliveryNotFound
		}

		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return del, nil
}

func (d *Deliveries) GetFromEvent(ctx context.Context, command dto.GetDeliveriesFromEventCommand) ([]domain.Delivery, int, error) {
	const fn = "infrastructure.pg.Deliveries.GetFromEvent"

	offset := (command.Page - 1) * command.Limit

	rows, err := d.pool.Query(ctx, `
		SELECT id, endpoint_id, status, attempts, max_attempts, next_retry_at, created_at, last_response_code, last_error
		FROM deliveries
		WHERE event_id = $1
		ORDER BY created_at ASC
		LIMIT $2
		OFFSET $3`, command.EventID, command.Limit, offset)

	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", fn, err)
	}

	defer rows.Close()

	var results []domain.Delivery
	for rows.Next() {
		r := domain.Delivery{EventID: command.EventID}
		if err := rows.Scan(&r.ID, &r.EndpointID, &r.Status, &r.Attempts, &r.MaxAttempts, &r.NextRetryAt, &r.CreatedAt, &r.LastResponseCode, &r.LastError); err != nil {
			return nil, 0, fmt.Errorf("%s: %w", fn, err)
		}

		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("%s: %w", fn, err)
	}

	var total int
	if err = d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM deliveries WHERE event_id = $1`, command.EventID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("%s: %w", fn, err)
	}

	return results, total, nil
}
