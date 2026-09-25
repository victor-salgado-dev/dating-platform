package favorites

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

func (r *PostgresRepository) Add(ctx context.Context, userID, profileID uuid.UUID) error {
	const query = `
		INSERT INTO favorites (user_id, favorite_profile_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, favorite_profile_id) DO NOTHING
	`

	if _, err := r.db.Exec(ctx, query, userID, profileID); err != nil {
		return fmt.Errorf("favorites: añadir favorito: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Remove(ctx context.Context, userID, profileID uuid.UUID) error {
	const query = `DELETE FROM favorites WHERE user_id = $1 AND favorite_profile_id = $2`

	if _, err := r.db.Exec(ctx, query, userID, profileID); err != nil {
		return fmt.Errorf("favorites: quitar favorito: %w", err)
	}
	return nil
}

func (r *PostgresRepository) IsFavorited(ctx context.Context, userID, profileID uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM favorites WHERE user_id = $1 AND favorite_profile_id = $2)`

	var exists bool
	if err := r.db.QueryRow(ctx, query, userID, profileID).Scan(&exists); err != nil {
		return false, fmt.Errorf("favorites: comprobar favorito: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepository) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	const query = `
		SELECT
			p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
			EXISTS (SELECT 1 FROM profile_photos ph WHERE ph.profile_id = p.id) AS has_photo,
			p.relationship_goals[1], f.created_at,
			COUNT(*) OVER() AS total_count
		FROM favorites f
		JOIN profiles p ON p.id = f.favorite_profile_id
		JOIN users u ON u.id = p.user_id
		WHERE f.user_id = $1 AND u.status = 'active' AND u.deleted_at IS NULL
		  AND NOT EXISTS (
		      SELECT 1 FROM blocks b
		      WHERE (b.blocker_id = $1 AND b.blocked_id = p.user_id)
		         OR (b.blocker_id = p.user_id AND b.blocked_id = $1)
		  )
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("favorites: listar favoritos: %w", err)
	}
	defer rows.Close()

	var items []ListItem
	total := 0

	for rows.Next() {
		var (
			relGoalStr  *string
			favoritedAt time.Time
			totalCount  int
		)

		base, scanErr := profiles.ScanBaseListItem(rows, &relGoalStr, &favoritedAt, &totalCount)
		if scanErr != nil {
			return nil, fmt.Errorf("favorites: leer favorito: %w", scanErr)
		}

		item := ListItem{
			BaseListItem: base,
			FavoritedAt:  favoritedAt,
		}
		if relGoalStr != nil {
			g := profiles.RelationshipGoal(*relGoalStr)
			item.RelationshipGoal = &g
		}

		total = totalCount
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("favorites: listar favoritos: %w", err)
	}

	return &ListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: pagination.TotalPages(total, pageSize),
	}, nil
}

func (r *PostgresRepository) ListReceived(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error) {
	const query = `
		SELECT
			p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
			EXISTS (SELECT 1 FROM profile_photos ph WHERE ph.profile_id = p.id) AS has_photo,
			p.relationship_goals[1], f.created_at,
			COUNT(*) OVER() AS total_count
		FROM favorites f
		JOIN profiles viewer ON viewer.id = $1
		JOIN profiles p ON p.user_id = f.user_id
		JOIN users u ON u.id = p.user_id
		WHERE f.favorite_profile_id = $1
		  AND u.status = 'active' AND u.deleted_at IS NULL
		  AND NOT EXISTS (
		      SELECT 1 FROM blocks b
		      WHERE (b.blocker_id = viewer.user_id AND b.blocked_id = p.user_id)
		         OR (b.blocker_id = p.user_id AND b.blocked_id = viewer.user_id)
		  )
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, profileID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("favorites: listar favoritos recibidos: %w", err)
	}
	defer rows.Close()

	var items []ListItem
	total := 0

	for rows.Next() {
		var (
			relGoalStr  *string
			favoritedAt time.Time
			totalCount  int
		)

		base, scanErr := profiles.ScanBaseListItem(rows, &relGoalStr, &favoritedAt, &totalCount)
		if scanErr != nil {
			return nil, fmt.Errorf("favorites: leer favorito recibido: %w", scanErr)
		}

		item := ListItem{
			BaseListItem: base,
			FavoritedAt:  favoritedAt,
		}
		if relGoalStr != nil {
			g := profiles.RelationshipGoal(*relGoalStr)
			item.RelationshipGoal = &g
		}

		total = totalCount
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("favorites: listar favoritos recibidos: %w", err)
	}

	return &ListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: pagination.TotalPages(total, pageSize),
	}, nil
}

func (r *PostgresRepository) ListMutual(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListResult, error) {
	// "me" resuelve mi perfil a partir de mi usuario: la mutualidad hay
	// que comprobarla por profile_id (favorites.favorite_profile_id
	// apunta a profiles.id), no por user_id.
	const query = `
		SELECT
			p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
			EXISTS (SELECT 1 FROM profile_photos ph WHERE ph.profile_id = p.id) AS has_photo,
			p.relationship_goals[1], f.created_at,
			COUNT(*) OVER() AS total_count
		FROM favorites f
		JOIN profiles me ON me.user_id = $1
		JOIN profiles p ON p.id = f.favorite_profile_id
		JOIN users u ON u.id = p.user_id
		WHERE f.user_id = $1
		  AND u.status = 'active' AND u.deleted_at IS NULL
		  AND EXISTS (
		      SELECT 1 FROM favorites back
		      WHERE back.user_id = p.user_id AND back.favorite_profile_id = me.id
		  )
		  AND NOT EXISTS (
		      SELECT 1 FROM blocks b
		      WHERE (b.blocker_id = $1 AND b.blocked_id = p.user_id)
		         OR (b.blocker_id = p.user_id AND b.blocked_id = $1)
		  )
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("favorites: listar favoritos mutuos: %w", err)
	}
	defer rows.Close()

	var items []ListItem
	total := 0

	for rows.Next() {
		var (
			relGoalStr  *string
			favoritedAt time.Time
			totalCount  int
		)

		base, scanErr := profiles.ScanBaseListItem(rows, &relGoalStr, &favoritedAt, &totalCount)
		if scanErr != nil {
			return nil, fmt.Errorf("favorites: leer favorito mutuo: %w", scanErr)
		}

		item := ListItem{
			BaseListItem: base,
			FavoritedAt:  favoritedAt,
		}
		if relGoalStr != nil {
			g := profiles.RelationshipGoal(*relGoalStr)
			item.RelationshipGoal = &g
		}

		total = totalCount
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("favorites: listar favoritos mutuos: %w", err)
	}

	return &ListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: pagination.TotalPages(total, pageSize),
	}, nil
}
