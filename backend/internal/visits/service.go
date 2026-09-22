package visits

import (
"context"
"fmt"

"github.com/google/uuid"

"dating-platform/backend/internal/profiles"
)

type ProfilesRepository interface {
GetByUserID(ctx context.Context, userID uuid.UUID) (*profiles.Profile, error)
}

type Service struct {
repo     Repository
profiles ProfilesRepository
}

func NewService(repo Repository, profiles ProfilesRepository) *Service {
return &Service{repo: repo, profiles: profiles}
}

func (s *Service) Record(ctx context.Context, userID, targetProfileID uuid.UUID) error {
p, err := s.profiles.GetByUserID(ctx, userID)
if err != nil {
return fmt.Errorf("visits: resolver perfil visitante: %w", err)
}
if p.ID == targetProfileID {
return ErrCannotVisitSelf
}
return s.repo.Record(ctx, p.ID, targetProfileID)
}

func (s *Service) ListSent(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
p, err := s.profiles.GetByUserID(ctx, userID)
if err != nil {
return nil, fmt.Errorf("visits: resolver perfil: %w", err)
}
return s.repo.ListSent(ctx, p.ID, page, pageSize)
}

func (s *Service) ListReceived(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
p, err := s.profiles.GetByUserID(ctx, userID)
if err != nil {
return nil, fmt.Errorf("visits: resolver perfil: %w", err)
}
return s.repo.ListReceived(ctx, p.ID, page, pageSize)
}
