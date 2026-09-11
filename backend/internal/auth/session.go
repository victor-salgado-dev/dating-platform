package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Session representa una sesión de usuario autenticado. El valor de ID
// es el que viaja en la cookie del navegador; en el almacén se guarda
// bajo su forma hasheada (ver hashToken), nunca en claro.
type Session struct {
	ID        string
	UserID    uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionStore persiste sesiones. Vive en Redis: las sesiones son datos
// transitorios con expiración, no la fuente de verdad de la cuenta
// (eso es la tabla `users` en PostgreSQL).
type SessionStore interface {
	// Create genera y persiste una nueva sesión para userID, con la
	// duración configurada (AuthConfig.SessionTTL). Devuelve el token
	// de sesión en claro, que se coloca en la cookie.
	Create(ctx context.Context, userID uuid.UUID) (token string, err error)

	// Get devuelve la sesión asociada a un token de cookie. Devuelve
	// ErrTokenInvalid si el token no existe o ha expirado.
	Get(ctx context.Context, token string) (*Session, error)

	// Delete invalida una sesión (logout).
	Delete(ctx context.Context, token string) error
}
