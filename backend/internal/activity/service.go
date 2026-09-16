package activity

import (
	"context"
	"sort"

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

func (s *Service) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*Result, error) {
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
	result, err := s.repo.List(ctx, profile.ID, page, pageSize)
	if err != nil {
		return nil, err
	}
	sortItems(result.Items)
	return result, nil
}

func sortItems(items []Item) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			if items[i].EventType == items[j].EventType {
				return items[i].ProfileID.String() < items[j].ProfileID.String()
			}
			return items[i].EventType < items[j].EventType
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}
