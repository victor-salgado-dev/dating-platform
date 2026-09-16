package likes

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Add(ctx context.Context, fromProfileID, toProfileID uuid.UUID) (bool, error)
	Remove(ctx context.Context, fromProfileID, toProfileID uuid.UUID) error
	IsLiked(ctx context.Context, fromProfileID, toProfileID uuid.UUID) (bool, error)
	HasMatch(ctx context.Context, profileA, profileB uuid.UUID) (bool, error)
	ListSent(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error)
	ListReceived(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error)
	ListMatches(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*MatchResult, error)
}
