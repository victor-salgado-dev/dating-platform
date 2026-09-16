package activity

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	List(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*Result, error)
}
