package visits

import (
	"time"

	"dating-platform/backend/internal/pagination"
	"dating-platform/backend/internal/profiles"
)

const DefaultPageSize = pagination.DefaultPageSize

type ListItem struct {
	profiles.BaseListItem
	RelationshipGoal *profiles.RelationshipGoal
	VisitedAt        time.Time
}

type ListResult struct {
	Items      []ListItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}
