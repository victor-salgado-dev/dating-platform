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

// SessionResolver es un camino rápido OPCIONAL para autenticar cada
// petición: RedisSessionStore lo implementa con un único viaje a Redis que
// resuelve la sesión, comprueba si la cuenta está bloqueada (suspendida o
// eliminada) y decide si toca refrescar last_active_at. Si el SessionStore
// no lo implementa, Service.Authenticate cae en CurrentUser (Redis + una
// consulta a Postgres por petición).
type SessionResolver interface {
	// Resolve devuelve el usuario de la sesión. touch es true como mucho una
	// vez cada pocos minutos por usuario: es la señal para escribir
	// users.last_active_at. Devuelve ErrTokenInvalid o ErrAccountSuspended.
	Resolve(ctx context.Context, token string) (userID uuid.UUID, touch bool, err error)
}

// Motivos de bloqueo de una cuenta en el almacén de sesiones.
const (
	BlockSuspended = "suspended"
	BlockDeleted   = "deleted"
)

// AccountBlocker corta el acceso de una cuenta de inmediato SIN tener que
// consultar Postgres en cada petición: se guarda una marca con la misma
// duración que las sesiones, así todas las sesiones anteriores quedan
// cubiertas. Lo implementa RedisSessionStore.
type AccountBlocker interface {
	BlockUser(ctx context.Context, userID uuid.UUID, reason string) error
	UnblockUser(ctx context.Context, userID uuid.UUID) error
}

// AccountSummary es lo que el header necesita de la cuenta y que /auth/me
// devuelve junto al usuario.
type AccountSummary struct {
	ProfileID  *uuid.UUID
	PhotoURL   *string
	Completion int // 0-100
	HasProfile bool
}

// SummaryProvider calcula el AccountSummary. Lo implementa
// profiles.SummaryStore (auth no importa profiles).
type SummaryProvider interface {
	AccountSummary(ctx context.Context, userID uuid.UUID) (AccountSummary, error)
}
