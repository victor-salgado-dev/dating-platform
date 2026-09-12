package reports

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Create inserta un reporte nuevo con status inicial 'pending'.
	Create(ctx context.Context, reporterID, reportedID uuid.UUID, reason Reason, description *string) error
}
