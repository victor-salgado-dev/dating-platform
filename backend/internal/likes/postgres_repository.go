package likes

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"dating-platform/backend/internal/profiles"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{db: db} }

var _ Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Add(ctx context.Context, fromProfileID, toProfileID uuid.UUID) (bool, error) {
	profileOne, profileTwo := orderPair(fromProfileID, toProfileID)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("likes: iniciar transacción: %w", err)
	}
	defer tx.Rollback(ctx)

	// Serializa las dos direcciones de una pareja para que dos likes
	// simultáneos no puedan omitir la creación del match.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1::text || ':' || $2::text, 0))`, profileOne, profileTwo); err != nil {
		return false, fmt.Errorf("likes: bloquear pareja: %w", err)
	}

	if _, err := tx.Exec(ctx, `
        INSERT INTO likes (from_profile_id, to_profile_id)
        VALUES ($1, $2)
        ON CONFLICT (from_profile_id, to_profile_id) DO NOTHING
    `, fromProfileID, toProfileID); err != nil {
		return false, fmt.Errorf("likes: añadir like: %w", err)
	}

	if _, err := tx.Exec(ctx, `
        INSERT INTO matches (profile_one_id, profile_two_id)
		SELECT LEAST($1::uuid, $2::uuid), GREATEST($1::uuid, $2::uuid)
        WHERE EXISTS (
            SELECT 1 FROM likes
			WHERE from_profile_id = $2::uuid AND to_profile_id = $1::uuid
        )
        ON CONFLICT (profile_one_id, profile_two_id) DO NOTHING
    `, fromProfileID, toProfileID); err != nil {
		return false, fmt.Errorf("likes: crear match: %w", err)
	}
	var matched bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM matches WHERE profile_one_id = $1::uuid AND profile_two_id = $2::uuid
		)
	`, profileOne, profileTwo).Scan(&matched); err != nil {
		return false, fmt.Errorf("likes: comprobar match: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("likes: confirmar transacción: %w", err)
	}
	return matched, nil
}

func (r *PostgresRepository) Remove(ctx context.Context, fromProfileID, toProfileID uuid.UUID) error {
	profileOne, profileTwo := orderPair(fromProfileID, toProfileID)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("likes: iniciar transacción: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1::text || ':' || $2::text, 0))`, profileOne, profileTwo); err != nil {
		return fmt.Errorf("likes: bloquear pareja: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM likes WHERE from_profile_id = $1 AND to_profile_id = $2`, fromProfileID, toProfileID); err != nil {
		return fmt.Errorf("likes: quitar like: %w", err)
	}
	if _, err := tx.Exec(ctx, `
        DELETE FROM matches
		WHERE profile_one_id = LEAST($1::uuid, $2::uuid)
		  AND profile_two_id = GREATEST($1::uuid, $2::uuid)
    `, fromProfileID, toProfileID); err != nil {
		return fmt.Errorf("likes: quitar match: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("likes: confirmar transacción: %w", err)
	}
	return nil
}

func (r *PostgresRepository) IsLiked(ctx context.Context, fromProfileID, toProfileID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM likes WHERE from_profile_id = $1 AND to_profile_id = $2)`, fromProfileID, toProfileID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("likes: comprobar like: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepository) HasMatch(ctx context.Context, profileA, profileB uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM matches
			WHERE profile_one_id = LEAST($1::uuid, $2::uuid)
			  AND profile_two_id = GREATEST($1::uuid, $2::uuid)
		)
	`, profileA, profileB).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("likes: comprobar match: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepository) ListSent(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error) {
	return r.listLikes(ctx, profileID, true, page, pageSize)
}

func (r *PostgresRepository) ListReceived(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*ListResult, error) {
	return r.listLikes(ctx, profileID, false, page, pageSize)
}

func (r *PostgresRepository) listLikes(ctx context.Context, profileID uuid.UUID, sent bool, page, pageSize int) (*ListResult, error) {
	direction := "l.from_profile_id = $1"
	profileColumn := "l.to_profile_id"
	if !sent {
		direction = "l.to_profile_id = $1"
		profileColumn = "l.from_profile_id"
	}
	query := fmt.Sprintf(`
        SELECT p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
               p.relationship_goals[1], l.created_at,
               EXISTS (SELECT 1 FROM profile_photos ph WHERE ph.profile_id = p.id),
               COUNT(*) OVER()
        FROM likes l
        JOIN profiles viewer ON viewer.id = $1
        JOIN profiles p ON p.id = %s
        JOIN users u ON u.id = p.user_id
        WHERE %s AND u.status = 'active' AND u.deleted_at IS NULL
          AND NOT EXISTS (
              SELECT 1 FROM blocks b
              WHERE (b.blocker_id = viewer.user_id AND b.blocked_id = p.user_id)
                 OR (b.blocker_id = p.user_id AND b.blocked_id = viewer.user_id)
          )
        ORDER BY l.created_at DESC
        LIMIT $2 OFFSET $3
    `, profileColumn, direction)
	return r.scanList(ctx, query, profileID, page, pageSize, "listar likes")
}

func (r *PostgresRepository) ListMatches(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*MatchResult, error) {
	const query = `
        SELECT p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
               EXISTS (SELECT 1 FROM profile_photos ph WHERE ph.profile_id = p.id),
               m.created_at, c.id, COUNT(*) OVER()
        FROM matches m
        JOIN profiles viewer ON viewer.id = $1
        JOIN profiles p ON p.id = CASE WHEN m.profile_one_id = $1 THEN m.profile_two_id ELSE m.profile_one_id END
        JOIN users u ON u.id = p.user_id
        LEFT JOIN conversations c ON c.user_one_id = LEAST(viewer.user_id, p.user_id)
                                  AND c.user_two_id = GREATEST(viewer.user_id, p.user_id)
        WHERE (m.profile_one_id = $1 OR m.profile_two_id = $1)
          AND u.status = 'active' AND u.deleted_at IS NULL
          AND NOT EXISTS (
              SELECT 1 FROM blocks b
              WHERE (b.blocker_id = viewer.user_id AND b.blocked_id = p.user_id)
                 OR (b.blocker_id = p.user_id AND b.blocked_id = viewer.user_id)
          )
        ORDER BY m.created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.Query(ctx, query, profileID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("likes: listar matches: %w", err)
	}
	defer rows.Close()
	now := time.Now()
	var items []MatchItem
	total := 0
	for rows.Next() {
		var item MatchItem
		var gender string
		var birth time.Time
		var totalCount int
		if err := rows.Scan(&item.ProfileID, &item.DisplayName, &birth, &gender, &item.CountryCode, &item.Region, &item.HasPhoto, &item.MatchedAt, &item.ConversationID, &totalCount); err != nil {
			return nil, fmt.Errorf("likes: leer match: %w", err)
		}
		item.Age = profiles.AgeAt(birth, now)
		item.Gender = profiles.Gender(gender)
		total = totalCount
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("likes: listar matches: %w", err)
	}
	return &MatchResult{Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages(total, pageSize)}, nil
}

func (r *PostgresRepository) scanList(ctx context.Context, query string, profileID uuid.UUID, page, pageSize int, action string) (*ListResult, error) {
	rows, err := r.db.Query(ctx, query, profileID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("likes: %s: %w", action, err)
	}
	defer rows.Close()
	now := time.Now()
	var items []ListItem
	total := 0
	for rows.Next() {
		var item ListItem
		var gender string
		var rel *string
		var birth time.Time
		var totalCount int
		if err := rows.Scan(&item.ProfileID, &item.DisplayName, &birth, &gender, &item.CountryCode, &item.Region, &rel, &item.LikedAt, &item.HasPhoto, &totalCount); err != nil {
			return nil, fmt.Errorf("likes: leer like: %w", err)
		}
		item.Age = profiles.AgeAt(birth, now)
		item.Gender = profiles.Gender(gender)
		if rel != nil {
			goal := profiles.RelationshipGoal(*rel)
			item.RelationshipGoal = &goal
		}
		total = totalCount
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("likes: %s: %w", action, err)
	}
	return &ListResult{Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages(total, pageSize)}, nil
}

func totalPages(total, pageSize int) int {
	if total == 0 {
		return 0
	}
	return (total + pageSize - 1) / pageSize
}

func orderPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	for i := range a {
		if a[i] < b[i] {
			return a, b
		}
		if a[i] > b[i] {
			return b, a
		}
	}
	return a, b
}
