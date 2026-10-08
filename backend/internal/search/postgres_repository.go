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

	// counts recuerda unos segundos el total de las búsquedas con filtros.
	counts *countCache

	// viewerScopes evita releer seeking_genders y las preferencias de edad en
	// cada búsqueda. ProfilesService invalida la entrada al guardar esos campos.
	viewerScopes *viewerScopeCache
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db:           db,
		presets:      newPresetCache(presetCacheTTL),
		counts:       newCountCache(countCacheTTL, countCacheMax),
		viewerScopes: newViewerScopeCache(viewerScopeCacheTTL, viewerScopeCacheMax),
	}
}

var _ Repository = (*PostgresRepository)(nil)

// totalFromPage deduce el total sin consultar la base de datos cuando la página
// devuelta no viene llena. El bool indica si el total es exacto.
func totalFromPage(params Params, got int) (int, bool) {
	if got < params.PageSize && (got > 0 || params.Page == 1) {
		return (params.Page-1)*params.PageSize + got, true
	}
	return 0, false
}

// countProfiles devuelve cuántos perfiles cumplen los filtros, recordándolo
// unos segundos para que paginar no repita el recuento.
func (r *PostgresRepository) countProfiles(ctx context.Context, q builtQuery) (int, error) {
	key := countKey(q)
	return r.counts.getOrLoad(ctx, key, func() (int, error) {
		var total int
		if err := r.db.QueryRow(ctx, q.CountSQL, q.CountArgs...).Scan(&total); err != nil {
			return 0, fmt.Errorf("search: contar perfiles: %w", err)
		}
		return total, nil
	})
}

// loadCards carga los datos de tarjeta de los ids de una página, en ese mismo
// orden, con la relación (like/favorito) respecto a quien mira. Un id que haya
// desaparecido entre las dos fases simplemente se omite.
func (r *PostgresRepository) loadCards(ctx context.Context, viewerUserID uuid.UUID, ids []uuid.UUID, now time.Time) ([]ResultItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(`
		SELECT
			p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
			COALESCE(p.relationship_goals, '{}'), p.created_at,
			(SELECT ph.id FROM profile_photos ph
				WHERE ph.profile_id = p.id
				ORDER BY ph.position ASC, ph.id ASC
				LIMIT 1),
			%s
		FROM profiles p
		LEFT JOIN profiles me ON me.user_id = $1
		WHERE p.id = ANY($2)
	`, profiles.ViewerFlagsSQL("me", "p"))

	rows, err := r.db.Query(ctx, query, viewerUserID, ids)
	if err != nil {
		return nil, fmt.Errorf("search: cargar tarjetas: %w", err)
	}
	defer rows.Close()

	byID := make(map[uuid.UUID]ResultItem, len(ids))
	for rows.Next() {
		var (
			item         ResultItem
			genderStr    string
			relGoalsList []string
			birthDate    time.Time
		)
		if err := rows.Scan(
			&item.ProfileID, &item.DisplayName, &birthDate, &genderStr, &item.CountryCode,
			&item.Region, &relGoalsList, &item.CreatedAt, &item.PhotoID,
			&item.Liked, &item.Favorited, &item.ReceivedLike, &item.ReceivedFavorite,
		); err != nil {
			return nil, fmt.Errorf("search: leer tarjeta: %w", err)
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
		byID[item.ProfileID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search: leer tarjetas: %w", err)
	}

	items := make([]ResultItem, 0, len(ids))
	for _, id := range ids {
		if it, ok := byID[id]; ok {
			items = append(items, it)
		}
	}
	return items, nil
}

// loadViewerScope lee el género buscado (profiles.seeking_genders) y, si
// withAge, el rango de edad de las preferencias de pareja de quien mira.
// Sin perfil o sin datos => sin restricción (nil = "no lo ha dicho").
func (r *PostgresRepository) loadViewerScope(ctx context.Context, userID uuid.UUID, withAge bool) (ViewerScope, error) {
	scope, err := r.viewerScopes.getOrLoad(ctx, userID, func() (ViewerScope, error) {
		return r.readViewerScope(ctx, userID)
	})
	if err != nil {
		return ViewerScope{}, err
	}
	if !withAge {
		scope.MinAge = nil
		scope.MaxAge = nil
	}
	return scope, nil
}

// InvalidateViewerScope se llama tras guardar los campos del perfil que
// condicionan la búsqueda (géneros y límites de edad preferidos).
func (r *PostgresRepository) InvalidateViewerScope(userID uuid.UUID) {
	if r.viewerScopes != nil {
		r.viewerScopes.invalidate(userID)
	}
}

func (r *PostgresRepository) readViewerScope(ctx context.Context, userID uuid.UUID) (ViewerScope, error) {
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
	if ageMin != nil {
		v := int(*ageMin)
		s.MinAge = &v
	}
	if ageMax != nil {
		v := int(*ageMax)
		s.MaxAge = &v
	}
	return s, nil
}

func (r *PostgresRepository) queryProfileIDs(ctx context.Context, query string, args ...any) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search: consultar perfiles: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("search: leer resultado: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search: leer resultados: %w", err)
	}
	return ids, nil
}

func (r *PostgresRepository) Search(ctx context.Context, params Params) (*Result, error) {
	now := time.Now()

	// Carga preferencias solo cuando la consulta las necesita. Un género enviado
	// explícitamente sustituye seeking_genders; un rango de edad explícito también
	// sustituye profile_partner_preferences aunque la ruta permita usarlo por defecto.
	needsViewerScope := len(params.Filters.Genders) == 0 ||
		(params.UseAgePrefs && params.Filters.MinAge == nil && params.Filters.MaxAge == nil)
	if needsViewerScope {
		viewer, err := r.loadViewerScope(ctx, params.ExcludeUserID, params.UseAgePrefs)
		if err != nil {
			return nil, err
		}
		params.Viewer = viewer
	}

	// Listados sin filtros: desde la caché en memoria cuando es posible.
	cached, ok, err := r.searchPreset(ctx, params, now)
	if err != nil {
		return nil, err
	}
	if ok {
		return cached, nil
	}

	q := buildSearchQuery(params, now)

	// Fase 1: solo los ids de la página. Es barata: no calcula likes, favoritos
	// ni foto, y como no cuenta, Postgres se detiene al juntar PageSize filas.
	ids, err := r.queryProfileIDs(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}
	pageLimit := params.PageSize
	if params.SkipTotal {
		pageLimit++
	}
	if q.FallbackSQL != "" && len(ids) < pageLimit {
		ids, err = r.queryProfileIDs(ctx, q.FallbackSQL, q.FallbackArgs...)
		if err != nil {
			return nil, err
		}
	}
	hasMore := false
	if params.SkipTotal && len(ids) > params.PageSize {
		hasMore = true
		ids = ids[:params.PageSize]
	}
	// Fase 2: datos de tarjeta y relación con quien mira, solo de esos ids.
	items, err := r.loadCards(ctx, params.ExcludeUserID, ids, now)
	if err != nil {
		return nil, err
	}

	// Total: si la página no viene llena, es exacto sin contar (todo lo que hay
	// son las páginas anteriores más estas filas). Si viene llena, o si la página
	// pedida está fuera de rango, hay que contar (con caché).
	total, totalPages := 0, 0
	totalExact := !params.SkipTotal
	if !params.SkipTotal {
		var exact bool
		total, exact = totalFromPage(params, len(ids))
		if !exact {
			total, err = r.countProfiles(ctx, q)
			if err != nil {
				return nil, err
			}
		}
		if total > 0 {
			totalPages = (total + params.PageSize - 1) / params.PageSize
		}
		hasMore = params.Page < totalPages
	}

	return &Result{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
		TotalExact: totalExact,
		HasMore:    hasMore,
	}, nil
}
