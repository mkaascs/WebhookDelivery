package endpoints

import (
	"context"
	"fmt"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func (r *Repo) Update(ctx context.Context, command dto.UpdateEndpointCommand) error {
	const fn = "repo.endpoints.Repo.Update"

	res, err := r.db.ExecContext(ctx, `
    	UPDATE endpoints SET
			url = COALESCE($1, url),
			is_active = COALESCE($2, is_active),
			description = COALESCE($3, description)
		WHERE id = $4`, command.URL, command.IsActive, command.Description, command.ID)

	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if affected == 0 {
		return domain.ErrEndpointNotFound
	}

	return nil
}
