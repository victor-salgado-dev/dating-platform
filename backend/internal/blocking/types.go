// Package blocking implementa el bloqueo entre usuarios: quién ha
// bloqueado a quién, y la comprobación IsBlocked que consumen otros
// módulos (profiles, search, favorites, messaging) para aplicar
// visibilidad mutua. Depende de profiles.Repository (vía GetByIDAny,
// que deliberadamente no aplica reglas de visibilidad) para resolver
// profile_id -> user_id; profiles no depende de blocking.
package blocking

import (
	"time"

	"dating-platform/backend/internal/pagination"
	"dating-platform/backend/internal/profiles"
)

const (
	DefaultPageSize = pagination.DefaultPageSize
	MaxPageSize     = pagination.MaxPageSize
)

// ListItem es la ficha resumida de una persona bloqueada.
type ListItem struct {
	profiles.BaseListItem
	BlockedAt time.Time
}

type ListResult struct {
	Items      []ListItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}
