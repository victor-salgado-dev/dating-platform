package activity

import (
	"context"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

type Service struct {
	repo Repository
	ids  *profiles.IDResolver
}

func NewService(repo Repository, ids *profiles.IDResolver) *Service {
	return &Service{repo: repo, ids: ids}
}

// List devuelve el feed de actividad. El repositorio ya lo entrega ordenado
// (created_at DESC, event_type, profile id); antes se reordenaba aquí en Go
// con exactamente el mismo criterio.
func (s *Service) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*Result, error) {
	profileID, err := s.ids.ProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return s.repo.List(ctx, profileID, page, pageSize)
}
