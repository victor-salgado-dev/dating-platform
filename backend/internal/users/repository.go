package users

import (
	"context"

	"github.com/google/uuid"
)

// Repository define las operaciones de persistencia del dominio users.
// Se define como interfaz para que los módulos que la consumen (p. ej.
// auth en la Fase 3) dependan de un contrato y no de PostgreSQL
// directamente; también facilita sustituirla en tests.
type Repository interface {
	// Create inserta una nueva cuenta. Rellena en u los campos generados
	// por la base de datos (ID, Status, CreatedAt, UpdatedAt).
	// Devuelve ErrDuplicateEmail si el email ya está en uso.
	Create(ctx context.Context, u *User) error

	// GetByID devuelve una cuenta activa (no eliminada) por su ID.
	// Devuelve ErrNotFound si no existe.
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail devuelve una cuenta activa (no eliminada) por su email
	// (comparación insensible a mayúsculas/minúsculas).
	// Devuelve ErrNotFound si no existe.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// SoftDelete marca una cuenta como eliminada (deleted_at + status)
	// sin borrar la fila físicamente, preservando el histórico.
	// Devuelve ErrNotFound si la cuenta no existe o ya estaba eliminada.
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
