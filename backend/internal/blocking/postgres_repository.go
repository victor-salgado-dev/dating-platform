package blocking

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"dating-platform/backend/internal/pagination"
	"dating-platform/backend/internal/profiles"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

var _ Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Add(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	const query = `
		INSERT INTO blocks (blocker_id, blocked_id)
		VALUES ($1, $2)
		ON CONFLICT (blocker_id, blocked_id) DO NOTHING
	`

	if _, err := r.db.Exec(ctx, query, blockerID, blockedID); err != nil {
		return fmt.Errorf("blocking: bloquear: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Remove(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	const query = `DELETE FROM blocks WHERE blocker_id = $1 AND blocked_id = $2`

	if _, err := r.db.Exec(ctx, query, blockerID, blockedID); err != nil {
		return fmt.Errorf("blocking: desbloquear: %w", err)
	}
	return nil
}

func (r *PostgresRepository) IsBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM blocks
			WHERE (blocker_id = $1 AND blocked_id = $2)
			   OR (blocker_id = $2 AND blocked_id = $1)
		)
	`

	var exists bool
	if err := r.db.QueryRow(ctx, query, userA, userB).Scan(&exists); err != nil {
		return false, fmt.Errorf("blocking: comprobar bloqueo: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepository) List(ctx context.Context, blockerID uuid.UUID, page, pageSize int) (*ListResult, error) {
	const query = `
		SELECT
			p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
			EXISTS (SELECT 1 FROM profile_photos ph WHERE ph.profile_id = p.id) AS has_photo,
			b.created_at,
			COUNT(*) OVER() AS total_count
		FROM blocks b
		JOIN profiles p ON p.user_id = b.blocked_id
		WHERE b.blocker_id = $1
		ORDER BY b.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, blockerID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("blocking: listar bloqueados: %w", err)
	}
	defer rows.Close()

	var items []ListItem
	total := 0

	for rows.Next() {
		var (
			blockedAt  time.Time
			totalCount int
		)

		base, scanErr := profiles.ScanBaseListItem(rows, &blockedAt, &totalCount)
		if scanErr != nil {
			return nil, fmt.Errorf("blocking: leer bloqueado: %w", scanErr)
		}

		items = append(items, ListItem{
			BaseListItem: base,
			BlockedAt:    blockedAt,
		})
		total = totalCount
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("blocking: listar bloqueados: %w", err)
	}

	return &ListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: pagination.TotalPages(total, pageSize),
	}, nil
}
