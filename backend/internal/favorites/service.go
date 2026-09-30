package favorites

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

// Add marca profileID como favorito de userID. Valida que el perfil
// exista y sea de una cuenta activa sin bloqueos (profiles.ErrNotFound
// cubre todos esos casos sin distinguir) y que no sea el propio perfil
// del usuario. Todo en una sola consulta ligera (IDResolver.ResolveTarget).
func (s *Service) Add(ctx context.Context, userID, profileID uuid.UUID) error {
	target, err := s.ids.ResolveTarget(ctx, userID, profileID)
	if err != nil {
		return err
	}
	if target.TargetUserID == userID {
		return ErrCannotFavoriteSelf
	}
	return s.repo.Add(ctx, userID, profileID)
}

// Remove quita profileID de los favoritos de userID. Es idempotente: no
// falla si no estaba en la lista.
func (s *Service) Remove(ctx context.Context, userID, profileID uuid.UUID) error {
	return s.repo.Remove(ctx, userID, profileID)
}

func (s *Service) IsFavorited(ctx context.Context, userID, profileID uuid.UUID) (bool, error) {
	return s.repo.IsFavorited(ctx, userID, profileID)
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.List(ctx, userID, page, pageSize)
}

func (s *Service) ListReceived(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	profileID, err := s.ids.ProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListReceived(ctx, profileID, page, pageSize)
}

// ListMutual lista los favoritos mutuos del usuario. El repositorio recibe
// el user_id directamente (la condición "yo marqué" se resuelve por user_id,
// y "me marcaron" se contrasta internamente contra mi profile_id).
func (s *Service) ListMutual(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListMutual(ctx, userID, page, pageSize)
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
