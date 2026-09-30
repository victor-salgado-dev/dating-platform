package visits

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// maxListPageSize acota page_size (antes el handler no lo limitaba y
// ?page_size=1000000 devolvía todo el historial de visitas).
const maxListPageSize = 50

// ProfileIDResolver resuelve el profile_id de un usuario sin cargar la fila
// completa del perfil. Lo satisface *profiles.IDResolver.
type ProfileIDResolver interface {
	ProfileID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

type Service struct {
	repo Repository
	ids  ProfileIDResolver
}

func NewService(repo Repository, ids ProfileIDResolver) *Service {
	return &Service{repo: repo, ids: ids}
}

// Record registra la visita de userID al perfil targetProfileID.
func (s *Service) Record(ctx context.Context, userID, targetProfileID uuid.UUID) error {
	own, err := s.ids.ProfileID(ctx, userID)
	if err != nil {
		return fmt.Errorf("visits: resolver perfil visitante: %w", err)
	}
	if own == targetProfileID {
		return ErrCannotVisitSelf
	}
	return s.repo.Record(ctx, own, targetProfileID)
}

func (s *Service) ListSent(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	own, err := s.ids.ProfileID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("visits: resolver perfil: %w", err)
	}
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListSent(ctx, own, page, pageSize)
}

func (s *Service) ListReceived(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	own, err := s.ids.ProfileID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("visits: resolver perfil: %w", err)
	}
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListReceived(ctx, own, page, pageSize)
}

func (s *Service) ListMutual(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	own, err := s.ids.ProfileID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("visits: resolver perfil: %w", err)
	}
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListMutual(ctx, own, page, pageSize)
}

func clampPaging(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > maxListPageSize {
		pageSize = maxListPageSize
	}
	return page, pageSize
}
