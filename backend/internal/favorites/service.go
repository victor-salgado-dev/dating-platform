package favorites

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

// Add marca profileID como favorito de userID. Valida que el perfil
// exista y sea de una cuenta activa (misma regla de privacidad básica
// que la Fase 6: profiles.ErrNotFound cubre ambos casos sin distinguir)
// y que no sea el propio perfil del usuario.
func (s *Service) Add(ctx context.Context, userID, profileID uuid.UUID) error {
	target, err := s.profiles.GetPublicByID(ctx, profileID, userID)
	if err != nil {
		return err
	}
	if target.UserID == userID {
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
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return s.repo.List(ctx, userID, page, pageSize)
}

func (s *Service) ListReceived(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	profile, err := s.profiles.GetByUserID(ctx, userID)
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
	return s.repo.ListReceived(ctx, profile.ID, page, pageSize)
}

// ListMutual lista los favoritos mutuos del usuario. A diferencia de
// ListReceived, no necesita resolver el perfil: el repositorio recibe el
// user_id directamente (la condición "yo marqué" se resuelve por user_id,
// y "me marcaron" se contrasta internamente contra mi profile_id).
func (s *Service) ListMutual(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return s.repo.ListMutual(ctx, userID, page, pageSize)
}
