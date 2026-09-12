package blocking

import (
	"context"

	"github.com/google/uuid"
)

// Repository persiste bloqueos.
//
// Add y Remove son idempotentes, igual que en favorites: bloquear dos
// veces o desbloquear algo que no estaba bloqueado no es un error.
type Repository interface {
	Add(ctx context.Context, blockerID, blockedID uuid.UUID) error
	Remove(ctx context.Context, blockerID, blockedID uuid.UUID) error

	// IsBlocked indica si hay un bloqueo entre userA y userB EN
	// CUALQUIER SENTIDO. Es la comprobación que consumen otros módulos
	// para aplicar visibilidad mutua.
	IsBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error)

	// List pagina a quién ha bloqueado blockerID (no al revés).
	List(ctx context.Context, blockerID uuid.UUID, page, pageSize int) (*ListResult, error)
}
