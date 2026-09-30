package activity

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

func (r *PostgresRepository) List(ctx context.Context, profileID uuid.UUID, page, pageSize int) (*Result, error) {
	query := fmt.Sprintf(`
        WITH events AS (
            SELECT 'like_received'::text AS event_type, l.from_profile_id AS profile_id, l.created_at
            FROM likes l
            WHERE l.to_profile_id = $1::uuid
            UNION ALL
            SELECT 'match_created'::text, CASE
                WHEN m.profile_one_id = $1::uuid THEN m.profile_two_id
                ELSE m.profile_one_id
            END, m.created_at
            FROM matches m
            WHERE m.profile_one_id = $1::uuid OR m.profile_two_id = $1::uuid
            UNION ALL
            SELECT 'favorite_received'::text, p.id, f.created_at
            FROM favorites f
            JOIN profiles p ON p.user_id = f.user_id
            WHERE f.favorite_profile_id = $1::uuid
        )
        SELECT e.event_type, p.id, p.display_name, p.birth_date, p.gender,
               p.country_code, p.region,
               EXISTS (SELECT 1 FROM profile_photos ph WHERE ph.profile_id = p.id),
               (
                   SELECT '/api/v1/profiles/' || p.id::text || '/photos/' || ph.id::text || '/file'
                   FROM profile_photos ph
                   WHERE ph.profile_id = p.id
                   ORDER BY ph.position ASC, ph.id ASC
                   LIMIT 1
               ) AS photo_url,
               e.created_at, COUNT(*) OVER(),
               %s
        FROM events e
        JOIN profiles viewer ON viewer.id = $1::uuid
        JOIN profiles p ON p.id = e.profile_id
        JOIN users u ON u.id = p.user_id
        WHERE u.status = 'active' AND u.deleted_at IS NULL
          AND NOT EXISTS (
              SELECT 1 FROM blocks b
              WHERE (b.blocker_id = viewer.user_id AND b.blocked_id = p.user_id)
                 OR (b.blocker_id = p.user_id AND b.blocked_id = viewer.user_id)
          )
        ORDER BY e.created_at DESC, e.event_type ASC, p.id ASC
        LIMIT $2 OFFSET $3
    `, profiles.ViewerFlagsSQL("viewer", "p"))

	rows, err := r.db.Query(ctx, query, profileID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("activity: listar actividad: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	var items []Item
	total := 0
	for rows.Next() {
		var item Item
		var birthDate time.Time
		var gender string
		var totalCount int
		if err := rows.Scan(
			&item.EventType,
			&item.ProfileID,
			&item.DisplayName,
			&birthDate,
			&gender,
			&item.CountryCode,
			&item.Region,
			&item.HasPhoto,
			&item.PhotoURL,
			&item.CreatedAt,
			&totalCount,
			&item.Liked,
			&item.Favorited,
			&item.ReceivedLike,
			&item.ReceivedFavorite,
		); err != nil {
			return nil, fmt.Errorf("activity: leer actividad: %w", err)
		}
		item.Age = profiles.AgeAt(birthDate, now)
		item.Gender = profiles.Gender(gender)
		total = totalCount
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("activity: listar actividad: %w", err)
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return &Result{Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages}, nil
}
