package search

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"dating-platform/backend/internal/profiles"
)

type PostgresRepository struct {
	db *pgxpool.Pool

	// presets sirve en memoria los listados sin filtros (new-members, popular).
	// Ver preset_cache.go. Si es nil, todo va por SQL.
	presets *presetCache
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db, presets: newPresetCache(presetCacheTTL)}
}

var _ Repository = (*PostgresRepository)(nil)

// loadViewerScope lee el género buscado (profiles.seeking_genders) y, si
// withAge, el rango de edad de las preferencias de pareja de quien mira.
// Sin perfil o sin datos => sin restricción (nil = "no lo ha dicho").
func (r *PostgresRepository) loadViewerScope(ctx context.Context, userID uuid.UUID, withAge bool) (ViewerScope, error) {
	var (
		genders        []string
		ageMin, ageMax *int16
	)
	err := r.db.QueryRow(ctx, `
		SELECT p.seeking_genders, pp.age_min, pp.age_max
		FROM profiles p
		LEFT JOIN profile_partner_preferences pp ON pp.profile_id = p.id
		WHERE p.user_id = $1`, userID).Scan(&genders, &ageMin, &ageMax)
	if errors.Is(err, pgx.ErrNoRows) {
		return ViewerScope{}, nil
	}
	if err != nil {
		return ViewerScope{}, fmt.Errorf("search: cargar preferencias de quien mira: %w", err)
	}

	var s ViewerScope
	for _, g := range genders {
		s.SeekingGenders = append(s.SeekingGenders, profiles.Gender(g))
	}
	if withAge {
		if ageMin != nil {
			v := int(*ageMin)
			s.MinAge = &v
		}
		if ageMax != nil {
			v := int(*ageMax)
			s.MaxAge = &v
		}
	}
	return s, nil
}

func (r *PostgresRepository) Search(ctx context.Context, params Params) (*Result, error) {
	now := time.Now()

	// Restricciones de quien mira (género buscado y, en recommended, edad).
	viewer, err := r.loadViewerScope(ctx, params.ExcludeUserID, params.UseAgePrefs)
	if err != nil {
		return nil, err
	}
	params.Viewer = viewer

	// Listados sin filtros: desde la caché en memoria cuando es posible.
	cached, ok, err := r.searchPreset(ctx, params, now)
	if err != nil {
		return nil, err
	}
	if ok {
		return cached, nil
	}

	q := buildSearchQuery(params, now)

	rows, err := r.db.Query(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, fmt.Errorf("search: consultar perfiles: %w", err)
	}
	defer rows.Close()

	var items []ResultItem
	total := 0

	for rows.Next() {
		var (
			item         ResultItem
			genderStr    string
			relGoalsList []string
			birthDate    time.Time
			totalCount   int
		)

		if err := rows.Scan(
			&item.ProfileID, &item.DisplayName, &birthDate, &genderStr, &item.CountryCode,
			&item.Region, &relGoalsList, &item.CreatedAt, &item.PhotoID,
			&item.Liked, &item.Favorited, &item.ReceivedLike, &item.ReceivedFavorite,
			&totalCount,
		); err != nil {
			return nil, fmt.Errorf("search: leer resultado: %w", err)
		}

		item.HasPhoto = item.PhotoID != nil
		item.Age = profiles.AgeAt(birthDate, now)
		item.Gender = profiles.Gender(genderStr)
		if len(relGoalsList) > 0 {
			item.RelationshipGoals = make([]profiles.RelationshipGoal, len(relGoalsList))
			for i, rg := range relGoalsList {
				item.RelationshipGoals[i] = profiles.RelationshipGoal(rg)
			}
		}

		total = totalCount
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search: leer resultados: %w", err)
	}

	// COUNT(*) OVER() vive en las filas devueltas: si la página pedida está
	// fuera de rango no hay ninguna y total saldría 0, con lo que el cliente
	// no sabría a qué página volver. En ese caso (y solo ese) se cuenta aparte.
	if len(items) == 0 && params.Page > 1 {
		if err := r.db.QueryRow(ctx, q.CountSQL, q.CountArgs...).Scan(&total); err != nil {
			return nil, fmt.Errorf("search: contar perfiles: %w", err)
		}
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + params.PageSize - 1) / params.PageSize
	}

	return &Result{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}
