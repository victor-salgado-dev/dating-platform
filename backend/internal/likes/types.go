package likes

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

type ListItem struct {
	profiles.BaseListItem
	RelationshipGoal *profiles.RelationshipGoal
	LikedAt          time.Time
	// PhotoID es el id de la foto principal (la de position más baja),
	// resuelto en la misma consulta que lista los likes para evitar una
	// consulta por ítem (N+1). Es nil si el perfil no tiene fotos.
	PhotoID *uuid.UUID
}

type MatchItem struct {
	profiles.BaseListItem
	MatchedAt      time.Time
	ConversationID *uuid.UUID
	// PhotoID es el id de la foto principal del otro perfil, resuelto en
	// la misma consulta que lista los matches (mismo patrón que en
	// discovery/search). Nil si no hay foto.
	PhotoID *uuid.UUID
}

type ListResult struct {
	Items      []ListItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type MatchResult struct {
	Items      []MatchItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}
