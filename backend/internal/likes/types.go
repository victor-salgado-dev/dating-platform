package likes

import (
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 50
)

type ListItem struct {
	ProfileID        uuid.UUID
	DisplayName      string
	Age              int
	Gender           profiles.Gender
	CountryCode      string
	Region           *string
	RelationshipGoal *profiles.RelationshipGoal
	HasPhoto         bool
	LikedAt          time.Time
}

type MatchItem struct {
	ProfileID      uuid.UUID
	DisplayName    string
	Age            int
	Gender         profiles.Gender
	CountryCode    string
	Region         *string
	HasPhoto       bool
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
