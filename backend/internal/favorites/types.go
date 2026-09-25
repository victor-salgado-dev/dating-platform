// Package favorites implementa la lista de favoritos de cada usuario:
// marcar/desmarcar un perfil como favorito y listarlos. Depende de
// profiles.Repository (interfaz) para validar contra qué perfiles se
// puede actuar, pero profiles no sabe nada de favorites: la dependencia
// va en un solo sentido, sin ciclos.
package favorites

import (
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/pagination"
	"dating-platform/backend/internal/profiles"
)

const (
	DefaultPageSize = pagination.DefaultPageSize
	MaxPageSize     = pagination.MaxPageSize
)

// ListItem es la ficha resumida de un perfil favorito, con el mismo nivel
// de detalle que un resultado de búsqueda (Fase 5): no expone
// bio/intereses/idiomas completos. Se reutiliza tanto para favoritos
// enviados y recibidos como para los mutuos.
type ListItem struct {
	profiles.BaseListItem
	RelationshipGoal *profiles.RelationshipGoal
	FavoritedAt      time.Time
	// PhotoID es el id de la foto principal (la de position más baja),
	// resuelto en la misma consulta que lista los favoritos para evitar
	// una consulta por ítem (N+1). Es nil si el perfil no tiene fotos.
	PhotoID *uuid.UUID
}

// ListResult es una página de favoritos.
type ListResult struct {
	Items      []ListItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}
