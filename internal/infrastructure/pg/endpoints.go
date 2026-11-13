package pg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Endpoints struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewEndpoints(pool *pgxpool.Pool, log *slog.Logger) *Endpoints {
	return &Endpoints{pool: pool, log: log}
}

func (e *Endpoints) Add(ctx context.Context, command dto.AddEndpointCommand) (*domain.Endpoint, error) {
	const fn = "infrastructure.pg.Endpoints.Add"
	log := e.log.With(slog.String("fn", fn))

	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	success := false
	defer func(tx pgx.Tx) {
		if !success {
			if err := tx.Rollback(ctx); err != nil {
				log.Warn("failed to rollback transaction", sloglib.Error(err))
			}
		}
	}(tx)

	endpoint := &domain.Endpoint{
		ID:          uuid.Must(uuid.NewV7()),
		URL:         command.URL,
		Secret:      command.Secret,
		EventTypes:  command.EventTypes,
		Description: command.Description,
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO endpoints(id, url, secret, description) 
		VALUES ($1, $2, $3, COALESCE($4, '')) 
		RETURNING is_active, created_at`,
		endpoint.ID, command.URL, command.Secret, command.Description).
		Scan(&endpoint.IsActive, &endpoint.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	if _, err := insertSubscription(ctx, tx, endpoint.ID, command.EventTypes); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	success = true
	return endpoint, nil
}

func (e *Endpoints) GetByID(ctx context.Context, id uuid.UUID) (*domain.Endpoint, error) {
	const fn = "infrastructure.pg.Endpoints.GetByID"

	endpoint := &domain.Endpoint{ID: id}

	err := e.pool.QueryRow(ctx, `
		SELECT url, secret, description, is_active, created_at
		FROM endpoints
		WHERE id = $1`, id).
		Scan(&endpoint.URL, &endpoint.Secret, &endpoint.Description, &endpoint.IsActive, &endpoint.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEndpointNotFound
		}

		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	subs, err := getEndpointSubscriptions(ctx, e.pool, endpoint.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	for _, sub := range subs {
		endpoint.EventTypes = append(endpoint.EventTypes, sub.EventType)
	}

	return endpoint, nil
}

func (e *Endpoints) GetAll(ctx context.Context, command dto.GetAllEndpointsCommand) ([]domain.Endpoint, int, error) {
	const fn = "infrastructure.pg.Endpoints.GetAll"

	offset := (command.Page - 1) * command.Limit

	rows, err := e.pool.Query(ctx, `
		SELECT id, url, secret, description, is_active, created_at
		FROM endpoints
		ORDER BY created_at ASC
		LIMIT $1
		OFFSET $2`, command.Limit, offset)

	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", fn, err)
	}

	defer rows.Close()

	results := make([]domain.Endpoint, 0)

	for rows.Next() {
		var ep domain.Endpoint
		if err := rows.Scan(&ep.ID, &ep.URL, &ep.Secret, &ep.Description, &ep.IsActive, &ep.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("%s: %w", fn, err)
		}

		subs, err := getEndpointSubscriptions(ctx, e.pool, ep.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", fn, err)
		}

		for _, sub := range subs {
			ep.EventTypes = append(ep.EventTypes, sub.EventType)
		}

		results = append(results, ep)
	}

	var total int
	err = e.pool.QueryRow(ctx, `SELECT COUNT(*) FROM endpoints`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", fn, err)
	}

	return results, total, nil
}

func (e *Endpoints) Delete(ctx context.Context, id uuid.UUID) error {
	const fn = "infrastructure.pg.Endpoints.Delete"

	res, err := e.pool.Exec(ctx, `DELETE FROM endpoints WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if res.RowsAffected() == 0 {
		return domain.ErrEndpointNotFound
	}

	return nil
}

func (e *Endpoints) Update(ctx context.Context, command dto.UpdateEndpointCommand) error {
	const fn = "infrastructure.pg.Endpoints.Update"

	res, err := e.pool.Exec(ctx, `
		UPDATE endpoints SET 
			url = COALESCE($1, url),
			is_active = COALESCE($2, is_active),
			description = COALESCE($3, description)
		WHERE id = $4`, command.URL, command.IsActive, command.Description, command.ID)

	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if res.RowsAffected() == 0 {
		return domain.ErrEndpointNotFound
	}

	return nil
}
