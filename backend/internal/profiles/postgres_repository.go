package profiles

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pgUniqueViolation = "23505"
	pgCheckViolation  = "23514"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

var _ Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Create(ctx context.Context, p *Profile) error {
	const query = `
		INSERT INTO profiles (
			user_id, display_name, birth_date, gender, country_code,
			region, languages, relationship_goal, has_children, wants_children, bio, interests
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		p.UserID, p.DisplayName, p.BirthDate, string(p.Gender), p.CountryCode,
		p.Region, p.Languages, relationshipGoalToDB(p.RelationshipGoal),
		p.HasChildren, p.WantsChildren, p.Bio, p.Interests,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return ErrAlreadyExists
			case pgCheckViolation:
				return invalidField(pgErr.ConstraintName, "no cumple las restricciones del perfil")
			}
		}
		return fmt.Errorf("profiles: crear perfil: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	const query = `
		SELECT id, user_id, display_name, birth_date, gender, country_code,
		       region, languages, relationship_goal, has_children, wants_children,
		       bio, interests, created_at, updated_at
		FROM profiles
		WHERE user_id = $1
	`

	return r.scanOne(ctx, query, userID)
}

func (r *PostgresRepository) GetPublicByID(ctx context.Context, id, viewerUserID uuid.UUID) (*Profile, error) {
	const query = `
		SELECT p.id, p.user_id, p.display_name, p.birth_date, p.gender, p.country_code,
		       p.region, p.languages, p.relationship_goal, p.has_children, p.wants_children,
		       p.bio, p.interests, p.created_at, p.updated_at
		FROM profiles p
		JOIN users u ON u.id = p.user_id
		WHERE p.id = $1
		  AND u.status = 'active' AND u.deleted_at IS NULL
		  AND NOT EXISTS (
		      SELECT 1 FROM blocks b
		      WHERE (b.blocker_id = $2 AND b.blocked_id = p.user_id)
		         OR (b.blocker_id = p.user_id AND b.blocked_id = $2)
		  )
	`

	return r.scanOne(ctx, query, id, viewerUserID)
}

func (r *PostgresRepository) GetByIDAny(ctx context.Context, id uuid.UUID) (*Profile, error) {
	const query = `
		SELECT id, user_id, display_name, birth_date, gender, country_code,
		       region, languages, relationship_goal, has_children, wants_children,
		       bio, interests, created_at, updated_at
		FROM profiles
		WHERE id = $1
	`

	return r.scanOne(ctx, query, id)
}

func (r *PostgresRepository) Update(ctx context.Context, userID uuid.UUID, patch ProfilePatch) (*Profile, error) {
	var setClauses []string
	var args []any
	argN := 1

	add := func(col string, val any) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, val)
		argN++
	}

	if patch.DisplayName != nil {
		add("display_name", *patch.DisplayName)
	}
	if patch.BirthDate != nil {
		add("birth_date", *patch.BirthDate)
	}
	if patch.Gender != nil {
		add("gender", string(*patch.Gender))
	}
	if patch.CountryCode != nil {
		add("country_code", *patch.CountryCode)
	}
	if patch.RegionSet {
		add("region", patch.Region)
	}
	if patch.LanguagesSet {
		add("languages", patch.Languages)
	}
	if patch.RelationshipGoalSet {
		add("relationship_goal", relationshipGoalToDB(patch.RelationshipGoal))
	}
	if patch.HasChildrenSet {
		add("has_children", patch.HasChildren)
	}
	if patch.WantsChildrenSet {
		add("wants_children", patch.WantsChildren)
	}
	if patch.BioSet {
		add("bio", patch.Bio)
	}
	if patch.InterestsSet {
		add("interests", patch.Interests)
	}

	if len(setClauses) == 0 {
		// Nada que actualizar: devolvemos el perfil tal cual está.
		return r.GetByUserID(ctx, userID)
	}

	query := fmt.Sprintf(
		"UPDATE profiles SET %s WHERE user_id = $%d",
		strings.Join(setClauses, ", "), argN,
	)
	args = append(args, userID)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgCheckViolation {
			return nil, invalidField(pgErr.ConstraintName, "no cumple las restricciones del perfil")
		}
		return nil, fmt.Errorf("profiles: actualizar perfil: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	return r.GetByUserID(ctx, userID)
}

func (r *PostgresRepository) scanOne(ctx context.Context, query string, args ...any) (*Profile, error) {
	var p Profile
	var genderStr string
	var relGoalStr *string

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&p.ID, &p.UserID, &p.DisplayName, &p.BirthDate, &genderStr, &p.CountryCode,
		&p.Region, &p.Languages, &relGoalStr, &p.HasChildren, &p.WantsChildren,
		&p.Bio, &p.Interests, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("profiles: consultar perfil: %w", err)
	}

	p.Gender = Gender(genderStr)
	if relGoalStr != nil {
		g := RelationshipGoal(*relGoalStr)
		p.RelationshipGoal = &g
	}

	return &p, nil
}

func relationshipGoalToDB(g *RelationshipGoal) *string {
	if g == nil {
		return nil
	}
	v := string(*g)
	return &v
}

// --- Fotos ------------------------------------------------------------

func (r *PostgresRepository) AddPhoto(ctx context.Context, profileID uuid.UUID, photo *Photo) error {
	const query = `
		INSERT INTO profile_photos (profile_id, storage_key, content_type, position)
		VALUES (
			$1, $2, $3,
			COALESCE((SELECT MAX(position) + 1 FROM profile_photos WHERE profile_id = $1), 0)
		)
		RETURNING id, position, created_at
	`

	err := r.db.QueryRow(ctx, query, profileID, photo.StorageKey, photo.ContentType).
		Scan(&photo.ID, &photo.Position, &photo.CreatedAt)
	if err != nil {
		return fmt.Errorf("profiles: guardar foto: %w", err)
	}

	photo.ProfileID = profileID
	return nil
}

func (r *PostgresRepository) ListPhotos(ctx context.Context, profileID uuid.UUID) ([]Photo, error) {
	const query = `
		SELECT id, profile_id, storage_key, content_type, position, created_at
		FROM profile_photos
		WHERE profile_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.Query(ctx, query, profileID)
	if err != nil {
		return nil, fmt.Errorf("profiles: listar fotos: %w", err)
	}
	defer rows.Close()

	var photos []Photo
	for rows.Next() {
		var ph Photo
		if err := rows.Scan(&ph.ID, &ph.ProfileID, &ph.StorageKey, &ph.ContentType, &ph.Position, &ph.CreatedAt); err != nil {
			return nil, fmt.Errorf("profiles: leer foto: %w", err)
		}
		photos = append(photos, ph)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("profiles: listar fotos: %w", err)
	}

	return photos, nil
}

func (r *PostgresRepository) CountPhotos(ctx context.Context, profileID uuid.UUID) (int, error) {
	const query = `SELECT COUNT(*) FROM profile_photos WHERE profile_id = $1`

	var count int
	if err := r.db.QueryRow(ctx, query, profileID).Scan(&count); err != nil {
		return 0, fmt.Errorf("profiles: contar fotos: %w", err)
	}
	return count, nil
}

func (r *PostgresRepository) GetPhoto(ctx context.Context, profileID, photoID uuid.UUID) (*Photo, error) {
	const query = `
		SELECT id, profile_id, storage_key, content_type, position, created_at
		FROM profile_photos
		WHERE id = $1 AND profile_id = $2
	`

	var ph Photo
	err := r.db.QueryRow(ctx, query, photoID, profileID).
		Scan(&ph.ID, &ph.ProfileID, &ph.StorageKey, &ph.ContentType, &ph.Position, &ph.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPhotoNotFound
		}
		return nil, fmt.Errorf("profiles: consultar foto: %w", err)
	}

	return &ph, nil
}

func (r *PostgresRepository) DeletePhoto(ctx context.Context, profileID, photoID uuid.UUID) error {
	const query = `DELETE FROM profile_photos WHERE id = $1 AND profile_id = $2`

	tag, err := r.db.Exec(ctx, query, photoID, profileID)
	if err != nil {
		return fmt.Errorf("profiles: borrar foto: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPhotoNotFound
	}

	return nil
}
