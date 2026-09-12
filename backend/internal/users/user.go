// Package users contiene el dominio de la CUENTA de usuario: identidad,
// credenciales y estado de la cuenta. No contiene datos de perfil público
// (eso vive en el módulo `profiles`, a partir de la Fase 4).
package users

import (
	"time"

	"github.com/google/uuid"
)

// Status representa el estado de una cuenta.
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
)

// Role representa el nivel de acceso de una cuenta. Necesario a partir
// de la Fase 10 (panel de administración).
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User es la entidad de dominio que representa una cuenta.
//
// PasswordHash nunca debe serializarse hacia el exterior (API, logs).
// La lógica de hashing/verificación de contraseñas se implementa en la
// Fase 3 (Autenticación); este módulo solo modela y persiste el dato.
type User struct {
	ID              uuid.UUID
	Email           string
	PasswordHash    string
	Status          Status
	Role            Role
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// IsActive indica si la cuenta puede operar con normalidad.
func (u *User) IsActive() bool {
	return u.Status == StatusActive && u.DeletedAt == nil
}

// IsEmailVerified indica si el email de la cuenta ha sido verificado.
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// IsAdmin indica si la cuenta tiene privilegios de administración.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
