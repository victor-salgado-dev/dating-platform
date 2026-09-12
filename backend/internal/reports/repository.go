package reports

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Create inserta un reporte nuevo con status inicial 'pending'.
	Create(ctx context.Context, reporterID, reportedID uuid.UUID, reason Reason, description *string) error

	// List pagina reportes para el panel de administración (Fase 10),
	// enriquecidos con el nombre de perfil de reportador/reportado
	// cuando existe. statusFilter es opcional.
	List(ctx context.Context, page, pageSize int, statusFilter *Status) (*ListResult, error)

	// UpdateStatus marca un reporte como revisado o descartado.
	// Devuelve ErrNotFound si no existe.
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
}
