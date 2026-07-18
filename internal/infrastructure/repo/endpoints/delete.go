package endpoints

import (
	"context"
	"fmt"
	"webhook-delivery/internal/domain"
)

func (r *Repo) Delete(ctx context.Context, id string) error {
	const fn = "repo.endpoints.Repo.Delete"

	res, err := r.db.ExecContext(ctx, `
		DELETE FROM endpoints WHERE id = $1`, id)

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
