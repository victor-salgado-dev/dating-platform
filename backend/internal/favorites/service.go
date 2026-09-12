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
