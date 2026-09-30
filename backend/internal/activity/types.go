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

	// PhotoURL es la URL de la foto principal ya armada por el repositorio
	// en la misma consulta que la lista (evita el N+1). Nil si el perfil no
	// tiene fotos.
	PhotoURL         *string
	HasPhoto         bool
	Liked            bool
	Favorited        bool
	ReceivedLike     bool
	ReceivedFavorite bool
	CreatedAt        time.Time
}

type Result struct {
	Items      []Item
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}