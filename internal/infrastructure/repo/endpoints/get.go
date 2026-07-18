package endpoints

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

func (r *Repo) GetByID(ctx context.Context, id string) (*domain.Endpoint, error) {
	const fn = "repo.endpoints.Repo.GetByID"
	log := r.log.With(slog.String("fn", fn))

	var endpoint domain.Endpoint
	err := r.db.QueryRowContext(ctx, `
    	SELECT id, url, secret, description, is_active, created_at 
    	FROM endpoints
    	WHERE id = $1`, id).Scan(&endpoint.ID, &endpoint.URL, &endpoint.Secret, &endpoint.Description, &endpoint.IsActive, &endpoint.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrEndpointNotFound
		}

		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	rows, err := r.db.QueryContext(ctx, `
    	SELECT event_type
    	FROM subscriptions
    	WHERE endpoint_id = $1`, id)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Warn("failed to close sql rows", sloglib.Error(err))
		}
	}(rows)

	for rows.Next() {
		var eventType string
		if err := rows.Scan(&eventType); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		endpoint.EventTypes = append(endpoint.EventTypes, eventType)
	}

	return &endpoint, rows.Err()
}

func (r *Repo) GetAll(ctx context.Context, command dto.GetAllEndpointsCommand) ([]domain.Endpoint, error) {
	const fn = "repo.endpoints.Repo.GetAll"
	log := r.log.With(slog.String("fn", fn))

	offset := (command.Page - 1) * command.Limit

	rows, err := r.db.QueryContext(ctx, `
    	SELECT id, url, secret, description, is_active, created_at
		FROM endpoints
		ORDER BY created_at
		LIMIT $1
		OFFSET $2`, command.Limit, offset)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	defer closeRowsAndLog(log, rows)

	var keys []string
	endpoints := make(map[string]domain.Endpoint)

	for rows.Next() {
		var endpoint domain.Endpoint
		err := rows.Scan(&endpoint.ID, &endpoint.URL, &endpoint.Secret, &endpoint.Description, &endpoint.IsActive, &endpoint.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		endpoints[endpoint.ID] = endpoint
		keys = append(keys, endpoint.ID)
	}

	if len(keys) == 0 {
		return []domain.Endpoint{}, nil
	}

	rows, err = r.db.QueryContext(ctx, `
    	SELECT endpoint_id, event_type
    	FROM subscriptions
    	WHERE endpoint_id = ANY($1)`, keys)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	defer closeRowsAndLog(log, rows)

	for rows.Next() {
		var endpointID, eventType string
		if err := rows.Scan(&endpointID, &eventType); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		endpoint, ok := endpoints[endpointID]
		if !ok {
			continue
		}

		endpoint.EventTypes = append(endpoint.EventTypes, eventType)
		endpoints[endpointID] = endpoint
	}

	var results []domain.Endpoint
	for _, id := range keys {
		results = append(results, endpoints[id])
	}

	return results, nil
}

func closeRowsAndLog(log *slog.Logger, rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		log.Warn("failed to close sql rows", sloglib.Error(err))
	}
}
