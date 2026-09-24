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
}

type MatchItem struct {
	profiles.BaseListItem
	MatchedAt      time.Time
	ConversationID *uuid.UUID
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
