package visits

import (
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/pagination"
	"dating-platform/backend/internal/profiles"
)

const DefaultPageSize = pagination.DefaultPageSize

type ListItem struct {
	profiles.BaseListItem
	RelationshipGoal *profiles.RelationshipGoal
	VisitedAt        time.Time
	// PhotoID es el id de la foto principal (la de position más baja),
	// resuelto en la misma consulta que lista las visitas para evitar una
	// consulta por ítem (N+1). Es nil si el perfil no tiene fotos.
	PhotoID *uuid.UUID
}

type ListResult struct {
	Items      []ListItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}
