package visits

import (
"context"

"github.com/google/uuid"
)

type Repository interface {
Record(ctx context.Context, visitorProfileID, visitedProfileID uuid.UUID) error
ListSent(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error)
ListReceived(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error)
ListMutual(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error)
}
