package consent

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Record inserta un consentimiento nuevo (nunca actualiza uno
	// existente: cada aceptación es una fila propia, con su fecha).
	Record(ctx context.Context, userID uuid.UUID, docType DocumentType, version string) error

	// ListByUser devuelve el historial de consentimientos de un
	// usuario, más recientes primero. Pensado para transparencia
	// (que la persona pueda ver qué aceptó y cuándo).
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Consent, error)
}
