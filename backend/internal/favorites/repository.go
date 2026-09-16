package favorites

import (
	"context"

	"github.com/google/uuid"
)

// Repository persiste la relación de favoritos.
//
// Add y Remove son idempotentes a propósito: marcar dos veces el mismo
// favorito, o desmarcar uno que ya no existe, no es un error — así el
// botón "favorito" del cliente no tiene que llevar la cuenta de en qué
// estado cree que está antes de llamar a la API.
type Repository interface {
	Add(ctx context.Context, userID, profileID uuid.UUID) error
	Remove(ctx context.Context, userID, profileID uuid.UUID) error
	IsFavorited(ctx context.Context, userID, profileID uuid.UUID) (bool, error)
	List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error)
	ListReceived(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error)
}
