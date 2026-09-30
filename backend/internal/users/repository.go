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

	// MarkEmailVerified fija email_verified_at a la hora actual.
	// Devuelve ErrNotFound si la cuenta no existe.
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error

	// UpdatePasswordHash sustituye el hash de contraseña almacenado
	// (usado en el flujo de reset de contraseña).
	// Devuelve ErrNotFound si la cuenta no existe.
	UpdatePasswordHash(ctx context.Context, id uuid.UUID, passwordHash string) error

	// SetStatus cambia el estado de una cuenta (p. ej. suspender o
	// reactivar). No toca deleted_at: para eliminar una cuenta se usa
	// SoftDelete. Devuelve ErrNotFound si la cuenta no existe.
	SetStatus(ctx context.Context, id uuid.UUID, status Status) error

	// TouchLastActive fija last_active_at a la hora actual. Lo llama auth
	// como mucho una vez cada pocos minutos por usuario (alimenta
	// /search/online-now). No modifica updated_at (ver migración 000020).
	TouchLastActive(ctx context.Context, id uuid.UUID) error

	// List pagina cuentas para el panel de administración (Fase 10).
	// A diferencia de GetByID/GetByEmail, SÍ incluye cuentas eliminadas
	// (el admin necesita verlas para auditoría); statusFilter es
	// opcional.
	List(ctx context.Context, page, pageSize int, statusFilter *Status) (*ListResult, error)
}

// ListResult es una página de cuentas para el panel de administración.
type ListResult struct {
	Items      []User
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}
