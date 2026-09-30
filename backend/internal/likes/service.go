package likes

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

// Add da like a targetProfileID. Devuelve si el like ha creado (o ya había)
// un match. La comprobación de visibilidad del destinatario y la resolución
// de mi perfil son una sola consulta ligera.
func (s *Service) Add(ctx context.Context, userID, targetProfileID uuid.UUID) (bool, error) {
	t, err := s.ids.ResolveTarget(ctx, userID, targetProfileID)
	if err != nil {
		return false, err
	}
	if t.ViewerProfileID == targetProfileID {
		return false, ErrCannotLikeSelf
	}
	return s.repo.Add(ctx, t.ViewerProfileID, targetProfileID)
}

func (s *Service) Remove(ctx context.Context, userID, targetProfileID uuid.UUID) error {
	t, err := s.ids.ResolveTarget(ctx, userID, targetProfileID)
	if err != nil {
		return err
	}
	if t.ViewerProfileID == targetProfileID {
		return ErrCannotLikeSelf
	}
	return s.repo.Remove(ctx, t.ViewerProfileID, targetProfileID)
}

func (s *Service) Status(ctx context.Context, userID, targetProfileID uuid.UUID) (bool, bool, error) {
	t, err := s.ids.ResolveTarget(ctx, userID, targetProfileID)
	if err != nil {
		return false, false, err
	}
	liked, err := s.repo.IsLiked(ctx, t.ViewerProfileID, targetProfileID)
	if err != nil {
		return false, false, err
	}
	matched, err := s.repo.HasMatch(ctx, t.ViewerProfileID, targetProfileID)
	return liked, matched, err
}

// IsLiked y HasMatch los usa profiles.Service (LikeChecker) para el estado
// del visitante en el perfil completo.
func (s *Service) IsLiked(ctx context.Context, fromProfileID, toProfileID uuid.UUID) (bool, error) {
	return s.repo.IsLiked(ctx, fromProfileID, toProfileID)
}

func (s *Service) HasMatch(ctx context.Context, profileA, profileB uuid.UUID) (bool, error) {
	return s.repo.HasMatch(ctx, profileA, profileB)
}

func (s *Service) ListSent(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	own, err := s.ids.ProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListSent(ctx, own, page, pageSize)
}

func (s *Service) ListReceived(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	own, err := s.ids.ProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListReceived(ctx, own, page, pageSize)
}

func (s *Service) ListMatches(ctx context.Context, userID uuid.UUID, page, pageSize int) (*MatchResult, error) {
	own, err := s.ids.ProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListMatches(ctx, own, page, pageSize)
}

func clampPaging(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}
