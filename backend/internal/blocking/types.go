// Package blocking implementa el bloqueo entre usuarios: quién ha
// bloqueado a quién, y la comprobación IsBlocked que consumen otros
// módulos (profiles, search, favorites, messaging) para aplicar
// visibilidad mutua. Depende de profiles.Repository (vía GetByIDAny,
// que deliberadamente no aplica reglas de visibilidad) para resolver
// profile_id -> user_id; profiles no depende de blocking.
package blocking

import (
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 50
)

// ListItem es la ficha resumida de una persona bloqueada.
type ListItem struct {
	ProfileID   uuid.UUID
	DisplayName string
	Age         int
	Gender      profiles.Gender
	CountryCode string
	Region      *string
	BlockedAt   time.Time
}

type ListResult struct {
	Items      []ListItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}
