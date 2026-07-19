package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	"webhook-delivery/internal/lib/uuid"
)

type Events struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewEvents(pool *pgxpool.Pool, log *slog.Logger) *Events {
	return &Events{pool: pool, log: log}
}

func (e *Events) Add(ctx context.Context, command dto.PublishEventCommand) (*domain.Event, error) {
	const fn = "infrastructure.pg.Events.Add"

	event := &domain.Event{
		Payload: command.Payload,
		Type:    command.Type,
	}

	err := e.pool.QueryRow(ctx, `
		INSERT INTO events (id, type, payload)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		uuid.NewEvent(), command.Type, command.Payload).
		Scan(&event.ID, &event.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return event, nil
}

func (e *Events) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	const fn = "infrastructure.pg.Events.GetByID"

	event := &domain.Event{ID: id}

	err := e.pool.QueryRow(ctx, `
		SELECT type, payload, created_at
		FROM events
		WHERE id = $1`, id).
		Scan(&event.Type, &event.Payload, &event.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEventNotFound
		}

		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return event, nil
}
