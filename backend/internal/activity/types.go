package activity

import (
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 50

	EventLikeReceived     = "like_received"
	EventMatchCreated     = "match_created"
	EventFavoriteReceived = "favorite_received"
)

type Item struct {
	EventType   string
	ProfileID   uuid.UUID
	DisplayName string
	Age         int
	Gender      profiles.Gender
	CountryCode string
	Region      *string
	PhotoURL    *string         `json:"photo_url"` 
	HasPhoto    bool
	CreatedAt   time.Time
}

type Result struct {
	Items      []Item
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}
