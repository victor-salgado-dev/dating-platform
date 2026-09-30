package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// TouchLastActive implementa Repository.TouchLastActive. Vive en su propio
// archivo para no reescribir postgres_repository.go entero: no cambia nada
// de lo que ya hay allí.
func (r *PostgresRepository) TouchLastActive(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE users
		SET last_active_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	if _, err := r.db.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("users: actualizar last_active_at: %w", err)
	}
	return nil
}
