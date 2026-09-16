package likes

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

func (s *Service) Add(ctx context.Context, userID, targetProfileID uuid.UUID) (bool, error) {
	target, err := s.profiles.GetPublicByID(ctx, targetProfileID, userID)
	if err != nil {
		return false, err
	}
	own, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	if own.ID == target.ID {
		return false, ErrCannotLikeSelf
	}
	return s.repo.Add(ctx, own.ID, target.ID)
}

func (s *Service) Remove(ctx context.Context, userID, targetProfileID uuid.UUID) error {
	target, err := s.profiles.GetPublicByID(ctx, targetProfileID, userID)
	if err != nil {
		return err
	}
	own, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if own.ID == target.ID {
		return ErrCannotLikeSelf
	}
	return s.repo.Remove(ctx, own.ID, target.ID)
}

func (s *Service) Status(ctx context.Context, userID, targetProfileID uuid.UUID) (bool, bool, error) {
	target, err := s.profiles.GetPublicByID(ctx, targetProfileID, userID)
	if err != nil {
		return false, false, err
	}
	own, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return false, false, err
	}
	liked, err := s.repo.IsLiked(ctx, own.ID, target.ID)
	if err != nil {
		return false, false, err
	}
	matched, err := s.repo.HasMatch(ctx, own.ID, target.ID)
	return liked, matched, err
}

func (s *Service) ListSent(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	own, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListSent(ctx, own.ID, page, pageSize)
}

func (s *Service) ListReceived(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	own, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListReceived(ctx, own.ID, page, pageSize)
}

func (s *Service) ListMatches(ctx context.Context, userID uuid.UUID, page, pageSize int) (*MatchResult, error) {
	own, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListMatches(ctx, own.ID, page, pageSize)
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
