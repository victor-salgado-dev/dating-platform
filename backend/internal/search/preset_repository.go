package search

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

// searchPreset sirve un listado sin filtros desde la caché en memoria. El bool
// es false cuando la petición no es cacheable (tiene filtros, orden distinto o
// hay demasiados perfiles): el llamador debe usar entonces la consulta SQL de
// siempre. El resultado es el mismo que daría esa consulta, salvo por el
// retraso de hasta presetCacheTTL en reflejar perfiles nuevos o cambios de foto.
func (r *PostgresRepository) searchPreset(ctx context.Context, params Params, now time.Time) (*Result, bool, error) {
	if r.presets == nil || !isPlainPreset(params) {
		return nil, false, nil
	}

	sort := params.Sort
	entry, err := r.presets.get(ctx, sort, func(ctx context.Context) (*presetEntry, error) {
		return r.buildPresetEntry(ctx, sort)
	})
	if err != nil {
		return nil, false, err
	}
	if entry == nil || !entry.complete {
		return nil, false, nil
	}

	// Siempre en vivo (no cacheado): así un bloqueo nuevo surte efecto al instante.
	excluded, err := r.excludedProfileIDs(ctx, params.ExcludeUserID)
	if err != nil {
		return nil, false, err
	}

	cards, total := pageOfCards(entry, excluded, params.Viewer, now, params.Page, params.PageSize)

	var items []ResultItem
	if len(cards) > 0 {
		ids := make([]uuid.UUID, len(cards))
		for i, c := range cards {
			ids[i] = c.ProfileID
		}

		flags, err := r.viewerFlagsFor(ctx, params.ExcludeUserID, ids)
		if err != nil {
			return nil, false, err
		}

		items = make([]ResultItem, 0, len(cards))
		for _, c := range cards {
			f, ok := flags[c.ProfileID]
			if !ok {
				// Dejó de ser visible (suspendido, eliminado, bloqueado) desde
				// que se cargó la lista: la revalidación de abajo lo descarta.
				continue
			}
			items = append(items, ResultItem{
				ProfileID:         c.ProfileID,
				DisplayName:       c.DisplayName,
				Age:               profiles.AgeAt(c.BirthDate, now),
				Gender:            c.Gender,
				CountryCode:       c.CountryCode,
				Region:            c.Region,
				RelationshipGoals: c.RelationshipGoals,
				HasPhoto:          c.PhotoID != nil,
				PhotoID:           c.PhotoID,
				CreatedAt:         c.CreatedAt,
				Liked:             f.liked,
				Favorited:         f.favorited,
				ReceivedLike:      f.receivedLike,
				ReceivedFavorite:  f.receivedFavorite,
			})
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
	}, true, nil
}

// buildPresetEntry carga la lista completa y ordenada de perfiles activos. La
// condición de actividad sale de profiles.VisibleSQL (única definición de la
// regla de visibilidad) con un usuario inexistente (uuid.Nil), de modo que no
// excluye a nadie por bloqueos: eso se hace después, por usuario.
func (r *PostgresRepository) buildPresetEntry(ctx context.Context, sort Sort) (*presetEntry, error) {
	joins := ""
	if sort == SortPopular {
		joins = "LEFT JOIN profile_popularity pop ON pop.profile_id = p.id"
	}

	query := fmt.Sprintf(`
		SELECT
			p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
			COALESCE(p.relationship_goals, '{}'), p.created_at,
			(SELECT ph.id FROM profile_photos ph
				WHERE ph.profile_id = p.id
				ORDER BY ph.position ASC, ph.id ASC
				LIMIT 1)
		FROM profiles p
		JOIN users u ON u.id = p.user_id
		%s
		WHERE %s
		ORDER BY %s
		LIMIT $2
	`, joins, profiles.VisibleSQL("$1"), orderByClause(sort))

	// Se pide uno más del máximo para saber si la lista cabe entera.
	rows, err := r.db.Query(ctx, query, uuid.Nil, presetCacheMaxProfiles+1)
	if err != nil {
		return nil, fmt.Errorf("search: cargar listado en caché: %w", err)
	}
	defer rows.Close()

	cards := make([]presetCard, 0, 1024)
	for rows.Next() {
		var (
			c         presetCard
			genderStr string
			goals     []string
		)
		if err := rows.Scan(
			&c.ProfileID, &c.DisplayName, &c.BirthDate, &genderStr, &c.CountryCode, &c.Region,
			&goals, &c.CreatedAt, &c.PhotoID,
		); err != nil {
			return nil, fmt.Errorf("search: leer listado en caché: %w", err)
		}

		c.Gender = profiles.Gender(genderStr)
		if len(goals) > 0 {
			c.RelationshipGoals = make([]profiles.RelationshipGoal, len(goals))
			for i, g := range goals {
				c.RelationshipGoals[i] = profiles.RelationshipGoal(g)
			}
		}
		cards = append(cards, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search: leer listado en caché: %w", err)
	}

	if len(cards) > presetCacheMaxProfiles {
		// No cabe: se deja constancia (sin guardar la lista) para no volver a
		// intentarlo en cada TTL, y las búsquedas siguen por SQL.
		return &presetEntry{complete: false, ttl: presetOversizeTTL}, nil
	}

	members := make(map[uuid.UUID]struct{}, len(cards))
	for _, c := range cards {
		members[c.ProfileID] = struct{}{}
	}
	return &presetEntry{cards: cards, members: members, complete: true}, nil
}

// excludedProfileIDs devuelve los perfiles que quien mira no debe ver en un
// listado por motivos propios de él: su propio perfil y los bloqueos en
// cualquiera de los dos sentidos. Tres búsquedas por índice, sin recorrer
// la tabla de perfiles.
func (r *PostgresRepository) excludedProfileIDs(ctx context.Context, viewerUserID uuid.UUID) ([]uuid.UUID, error) {
	const query = `
		SELECT p.id FROM profiles p WHERE p.user_id = $1
		UNION
		SELECT p.id FROM blocks b JOIN profiles p ON p.user_id = b.blocked_id WHERE b.blocker_id = $1
		UNION
		SELECT p.id FROM blocks b JOIN profiles p ON p.user_id = b.blocker_id WHERE b.blocked_id = $1
	`

	rows, err := r.db.Query(ctx, query, viewerUserID)
	if err != nil {
		return nil, fmt.Errorf("search: consultar exclusiones: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("search: leer exclusiones: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

type viewerFlags struct {
	liked, favorited, receivedLike, receivedFavorite bool
}

// viewerFlagsFor calcula, solo para los perfiles de la página, la relación con
// quien mira (misma definición que la consulta de búsqueda: ViewerFlagsSQL).
// Aprovecha para revalidar la visibilidad pública de esos perfiles con
// profiles.VisibleSQL: los que ya no la cumplan no aparecen en el mapa.
func (r *PostgresRepository) viewerFlagsFor(ctx context.Context, viewerUserID uuid.UUID, profileIDs []uuid.UUID) (map[uuid.UUID]viewerFlags, error) {
	query := fmt.Sprintf(`
		SELECT p.id, %s
		FROM profiles p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN profiles me ON me.user_id = $1
		WHERE p.id = ANY($2)
		  AND p.user_id <> $1
		  AND %s
	`, profiles.ViewerFlagsSQL("me", "p"), profiles.VisibleSQL("$1"))

	rows, err := r.db.Query(ctx, query, viewerUserID, profileIDs)
	if err != nil {
		return nil, fmt.Errorf("search: consultar relación con el perfil: %w", err)
	}
	defer rows.Close()

	out := make(map[uuid.UUID]viewerFlags, len(profileIDs))
	for rows.Next() {
		var (
			id uuid.UUID
			f  viewerFlags
		)
		if err := rows.Scan(&id, &f.liked, &f.favorited, &f.receivedLike, &f.receivedFavorite); err != nil {
			return nil, fmt.Errorf("search: leer relación con el perfil: %w", err)
		}
		out[id] = f
	}
	return out, rows.Err()
}
