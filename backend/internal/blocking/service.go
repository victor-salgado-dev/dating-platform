package blocking

import (
	"context"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

type Service struct {
	repo     Repository
	profiles profiles.Repository
}

func NewService(repo Repository, profilesRepo profiles.Repository) *Service {
	return &Service{repo: repo, profiles: profilesRepo}
}

// Add bloquea al usuario detrás de targetProfileID. Usa GetByIDAny (no
// GetPublicByID): bloquear tiene que funcionar incluso si esa persona
// ya te bloqueó a ti, o su cuenta está suspendida.
func (s *Service) Add(ctx context.Context, blockerUserID, targetProfileID uuid.UUID) error {
	target, err := s.profiles.GetByIDAny(ctx, targetProfileID)
	if err != nil {
		return err
	}
	if target.UserID == blockerUserID {
		return ErrCannotBlockSelf
	}
	return s.repo.Add(ctx, blockerUserID, target.UserID)
}

func (s *Service) Remove(ctx context.Context, blockerUserID, targetProfileID uuid.UUID) error {
	target, err := s.profiles.GetByIDAny(ctx, targetProfileID)
	if err != nil {
		return err
	}
	return s.repo.Remove(ctx, blockerUserID, target.UserID)
}

func (s *Service) IsBlockedByMe(ctx context.Context, blockerUserID, targetProfileID uuid.UUID) (bool, error) {
	target, err := s.profiles.GetByIDAny(ctx, targetProfileID)
	if err != nil {
		return false, err
	}
	return s.repo.IsBlocked(ctx, blockerUserID, target.UserID)
}

func (s *Service) IsBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	return s.repo.IsBlocked(ctx, userA, userB)
}

func (s *Service) List(ctx context.Context, blockerUserID uuid.UUID, page, pageSize int) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return s.repo.List(ctx, blockerUserID, page, pageSize)
}
