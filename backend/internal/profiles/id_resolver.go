package profiles

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// idResolverMaxEntries acota el mapa en memoria. Cuando se llena se vacía
// entero: es una caché de conveniencia, no una fuente de verdad.
const idResolverMaxEntries = 50_000

// IDResolver resuelve identificadores sin cargar la fila completa del perfil
// (unas 60 columnas). Los módulos de likes, favoritos, visitas y actividad
// solo necesitan el profile_id de quien actúa y, al mutar, saber si el
// destinatario es visible; antes lo hacían con GetByUserID y GetPublicByID.
type IDResolver struct {
	db *pgxpool.Pool

	mu    sync.RWMutex
	cache map[uuid.UUID]uuid.UUID // user_id -> profile_id
}

func NewIDResolver(db *pgxpool.Pool) *IDResolver {
	return &IDResolver{db: db, cache: make(map[uuid.UUID]uuid.UUID)}
}

// ProfileID devuelve el id del perfil de un usuario. La relación
// usuario-perfil es 1:1 e inmutable, por lo que solo se cachean los
// aciertos. Devuelve ErrNotFound si el usuario aún no tiene perfil.
func (r *IDResolver) ProfileID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	r.mu.RLock()
	id, ok := r.cache[userID]
	r.mu.RUnlock()
	if ok {
		return id, nil
	}

	err := r.db.QueryRow(ctx, `SELECT id FROM profiles WHERE user_id = $1`, userID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, fmt.Errorf("profiles: resolver profile_id: %w", err)
	}

	r.remember(userID, id)
	return id, nil
}

// Target es el resultado de ResolveTarget.
type Target struct {
	ViewerProfileID uuid.UUID // perfil de quien actúa
	TargetUserID    uuid.UUID // cuenta dueña del perfil destinatario
}

// ResolveTarget comprueba en UNA consulta lo que antes hacían GetPublicByID
// (destinatario existente, cuenta activa, sin bloqueos en ningún sentido) y
// GetByUserID (perfil de quien actúa). Como GetPublicByID, no distingue el
// motivo: cualquier fallo devuelve ErrNotFound.
func (r *IDResolver) ResolveTarget(ctx context.Context, viewerUserID, targetProfileID uuid.UUID) (Target, error) {
	const query = `
		SELECT me.id, p.user_id
		FROM profiles p
		JOIN users u ON u.id = p.user_id
		JOIN profiles me ON me.user_id = $2
		WHERE p.id = $1
		  AND ` + visiblePredicate + `
	`

	var t Target
	if err := r.db.QueryRow(ctx, query, targetProfileID, viewerUserID).Scan(&t.ViewerProfileID, &t.TargetUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Target{}, ErrNotFound
		}
		return Target{}, fmt.Errorf("profiles: resolver destinatario: %w", err)
	}

	r.remember(viewerUserID, t.ViewerProfileID)
	return t, nil
}

func (r *IDResolver) remember(userID, profileID uuid.UUID) {
	r.mu.Lock()
	if len(r.cache) >= idResolverMaxEntries {
		r.cache = make(map[uuid.UUID]uuid.UUID)
	}
	r.cache[userID] = profileID
	r.mu.Unlock()
}
