package endpoints

import (
	"context"
	"fmt"
	"log/slog"
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
	defer func() {
		if success {
			return
		}

		if err := tx.Rollback(); err != nil {
			log.Error("failed to rollback tx", sloglib.Error(err))
		}
	}()

	subsID := make([]string, 0, len(command.EventTypes))
	for count := 0; count < len(command.EventTypes); count++ {
		subsID = append(subsID, uuid.NewSubscription())
	}

	return nil, nil
}
