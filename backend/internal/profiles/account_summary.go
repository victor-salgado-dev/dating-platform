package profiles

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"dating-platform/backend/internal/auth"
)

// SummaryStore calcula lo que el header necesita de la cuenta (avatar y
// porcentaje de perfil completado) en una sola consulta ligera, para que
// GET /auth/me lo devuelva y el cliente no tenga que pedir perfil y fotos
// aparte. Implementa auth.SummaryProvider.
type SummaryStore struct {
	db *pgxpool.Pool
}

func NewSummaryStore(db *pgxpool.Pool) *SummaryStore { return &SummaryStore{db: db} }

var _ auth.SummaryProvider = (*SummaryStore)(nil)

// Criterios (los mismos que calculaba account-nav.tsx en el cliente): 4
// puntos base por tener perfil, 1 por cada uno de region, objetivos,
// has_children, bio, wants_children y nationality (un texto vacío no cuenta
// como contestado), y 2 por tener al menos una foto. Máximo 12.
const (
	completionBase     = 4
	completionPhotoPts = 2
	completionMax      = 12
)

func (s *SummaryStore) AccountSummary(ctx context.Context, userID uuid.UUID) (auth.AccountSummary, error) {
	const query = `
		SELECT p.id,
		       (SELECT ph.id FROM profile_photos ph
		         WHERE ph.profile_id = p.id
		         ORDER BY ph.position ASC, ph.id ASC
		         LIMIT 1) AS photo_id,
		       (COALESCE(btrim(p.region), '') <> '')::int
		     + (COALESCE(cardinality(p.relationship_goals), 0) > 0)::int
		     + (p.has_children IS NOT NULL)::int
		     + (COALESCE(btrim(p.bio), '') <> '')::int
		     + (p.wants_children IS NOT NULL)::int
		     + (COALESCE(btrim(p.nationality), '') <> '')::int AS optional_points
		FROM profiles p
		WHERE p.user_id = $1
	`

	var (
		profileID uuid.UUID
		photoID   *uuid.UUID
		optional  int
	)
	if err := s.db.QueryRow(ctx, query, userID).Scan(&profileID, &photoID, &optional); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.AccountSummary{}, nil // sin perfil todavía
		}
		return auth.AccountSummary{}, fmt.Errorf("profiles: resumen de cuenta: %w", err)
	}

	sum := auth.AccountSummary{ProfileID: &profileID, HasProfile: true}
	if photoID != nil {
		// Es un avatar del header: basta la miniatura, no la foto de 1600 px.
		url := publicPhotoURL(profileID, *photoID, true)
		sum.PhotoURL = &url
	}
	sum.Completion = completionPercent(optional, photoID != nil)
	return sum, nil
}

// completionPercent convierte los puntos obtenidos en el porcentaje 0-100 que
// ve el usuario.
func completionPercent(optionalPoints int, hasPhoto bool) int {
	score := completionBase + optionalPoints
	if hasPhoto {
		score += completionPhotoPts
	}
	return int(math.Round(float64(score) / completionMax * 100))
}
