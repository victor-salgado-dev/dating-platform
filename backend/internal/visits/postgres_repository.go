package visits

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"dating-platform/backend/internal/pagination"
	"dating-platform/backend/internal/profiles"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

var _ Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Record(ctx context.Context, visitorProfileID, visitedProfileID uuid.UUID) error {
	const query = `
INSERT INTO profile_visits (visitor_profile_id, visited_profile_id, visited_at)
VALUES ($1, $2, now())
ON CONFLICT (visitor_profile_id, visited_profile_id)
DO UPDATE SET visited_at = now()
`
	if _, err := r.db.Exec(ctx, query, visitorProfileID, visitedProfileID); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23503" {
				return profiles.ErrNotFound
			}
			if pgErr.Code == "23514" {
				return ErrCannotVisitSelf
			}
		}
		return fmt.Errorf("visits: registrar visita: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListSent(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error) {
	return r.listVisits(ctx, profileID, true, page, pageSize)
}

func (r *PostgresRepository) ListReceived(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error) {
	return r.listVisits(ctx, profileID, false, page, pageSize)
}

func (r *PostgresRepository) ListMutual(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error) {
	const query = `
SELECT p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
       EXISTS (SELECT 1 FROM profile_photos ph WHERE ph.profile_id = p.id) AS has_photo,
       (SELECT ph.id FROM profile_photos ph
            WHERE ph.profile_id = p.id
            ORDER BY ph.position ASC, ph.id ASC
            LIMIT 1) AS photo_id,
       GREATEST(v.visited_at, r.visited_at) AS visited_at,
       COUNT(*) OVER()
FROM profile_visits v
JOIN profile_visits r
  ON r.visitor_profile_id = v.visited_profile_id
 AND r.visited_profile_id = v.visitor_profile_id
JOIN profiles viewer ON viewer.id = $1
JOIN profiles p ON p.id = v.visited_profile_id
JOIN users u ON u.id = p.user_id
WHERE v.visitor_profile_id = $1
  AND u.status = 'active' AND u.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM blocks b
      WHERE (b.blocker_id = viewer.user_id AND b.blocked_id = p.user_id)
         OR (b.blocker_id = p.user_id AND b.blocked_id = viewer.user_id)
  )
ORDER BY GREATEST(v.visited_at, r.visited_at) DESC
LIMIT $2 OFFSET $3
`

	rows, err := r.db.Query(ctx, query, profileID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("visits: listar visitas mutuas: %w", err)
	}
	defer rows.Close()

	var items []ListItem
	total := 0

	for rows.Next() {
		var (
			photoID    *uuid.UUID
			visitedAt  time.Time
			totalCount int
		)

		base, scanErr := profiles.ScanBaseListItem(rows, &photoID, &visitedAt, &totalCount)
		if scanErr != nil {
			return nil, fmt.Errorf("visits: leer visita mutua: %w", scanErr)
		}

		items = append(items, ListItem{
			BaseListItem: base,
			VisitedAt:    visitedAt,
			PhotoID:      photoID,
		})
		total = totalCount
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("visits: listar visitas mutuas: %w", err)
	}

	return &ListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: pagination.TotalPages(total, pageSize),
	}, nil
}

func (r *PostgresRepository) listVisits(ctx context.Context, profileID uuid.UUID, sent bool, page, pageSize int) (*ListResult, error) {
	direction := "v.visitor_profile_id = $1"
	profileColumn := "v.visited_profile_id"
	if !sent {
		direction = "v.visited_profile_id = $1"
		profileColumn = "v.visitor_profile_id"
	}
	query := fmt.Sprintf(`
SELECT p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
       EXISTS (SELECT 1 FROM profile_photos ph WHERE ph.profile_id = p.id) AS has_photo,
       (SELECT ph.id FROM profile_photos ph
            WHERE ph.profile_id = p.id
            ORDER BY ph.position ASC, ph.id ASC
            LIMIT 1) AS photo_id,
       v.visited_at,
       COUNT(*) OVER()
FROM profile_visits v
JOIN profiles viewer ON viewer.id = $1
JOIN profiles p ON p.id = %s
JOIN users u ON u.id = p.user_id
WHERE %s AND u.status = 'active' AND u.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM blocks b
      WHERE (b.blocker_id = viewer.user_id AND b.blocked_id = p.user_id)
         OR (b.blocker_id = p.user_id AND b.blocked_id = viewer.user_id)
  )
ORDER BY v.visited_at DESC
LIMIT $2 OFFSET $3
`, profileColumn, direction)

	rows, err := r.db.Query(ctx, query, profileID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("visits: listar visitas: %w", err)
	}
	defer rows.Close()

	var items []ListItem
	total := 0

	for rows.Next() {
		var (
			photoID    *uuid.UUID
			visitedAt  time.Time
			totalCount int
		)

		base, scanErr := profiles.ScanBaseListItem(rows, &photoID, &visitedAt, &totalCount)
		if scanErr != nil {
			return nil, fmt.Errorf("visits: leer visita: %w", scanErr)
		}

		items = append(items, ListItem{
			BaseListItem: base,
			VisitedAt:    visitedAt,
			PhotoID:      photoID,
		})
		total = totalCount
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("visits: listar visitas: %w", err)
	}

	return &ListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: pagination.TotalPages(total, pageSize),
	}, nil
}
