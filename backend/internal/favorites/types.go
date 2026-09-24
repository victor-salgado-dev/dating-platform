// Package favorites implementa la lista de favoritos de cada usuario:
// marcar/desmarcar un perfil como favorito y listarlos. Depende de
// profiles.Repository (interfaz) para validar contra qué perfiles se
// puede actuar, pero profiles no sabe nada de favorites: la dependencia
// va en un solo sentido, sin ciclos.
package favorites

import (
	"time"

	"dating-platform/backend/internal/pagination"
	"dating-platform/backend/internal/profiles"
)

const (
	DefaultPageSize = pagination.DefaultPageSize
	MaxPageSize     = pagination.MaxPageSize
)

// ListItem es la ficha resumida de un perfil favorito, con el mismo nivel
// de detalle que un resultado de búsqueda (Fase 5): no expone
// bio/intereses/idiomas completos.
type ListItem struct {
	profiles.BaseListItem
	RelationshipGoal *profiles.RelationshipGoal
	FavoritedAt      time.Time
}

// ListResult es una página de favoritos.
type ListResult struct {
	Items      []ListItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}
