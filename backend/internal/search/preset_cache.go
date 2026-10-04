package search

import (
	"context"
	"log/slog"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

// Caché en memoria de los listados "sin filtros" (new-members, popular y
// search sin criterios). Cambia CPU de Postgres por RAM del backend:
//
//   - Una vez cada presetCacheTTL se carga UNA lista ordenada con todos los
//     perfiles activos (datos de la tarjeta + foto principal). Es la parte cara
//     y es idéntica para todos los usuarios.
//   - En cada petición solo se resuelve lo propio de quien mira, con consultas
//     diminutas: qué perfiles tiene excluidos (él mismo y bloqueos en ambos
//     sentidos) y los indicadores liked/favorited de los 20 perfiles de la
//     página, que además se revalidan con la regla de visibilidad pública.
//
// Si hay más de presetCacheMaxProfiles perfiles activos la caché no se usa y
// todo sigue por la consulta SQL de siempre (ver searchPreset).
const (
	// presetCacheTTL es lo que tarda en reflejarse un perfil nuevo, una foto
	// nueva o un cambio de orden. Los bloqueos NO dependen de este valor: se
	// consultan en cada petición.
	presetCacheTTL = 15 * time.Second

	// presetCacheMaxProfiles acota la memoria: ~250 B por perfil y orden, es
	// decir, unos 12 MB por orden con 50.000 perfiles.
	presetCacheMaxProfiles = 50_000

	// presetOversizeTTL: cuando hay más perfiles de los que caben, se recuerda
	// durante este tiempo para no reintentar la carga completa en cada TTL.
	presetOversizeTTL = 5 * time.Minute

	presetBuildTimeout = 15 * time.Second
)

// presetCard es la parte de una tarjeta que no depende de quien mira.
type presetCard struct {
	ProfileID         uuid.UUID
	DisplayName       string
	BirthDate         time.Time
	Gender            profiles.Gender
	CountryCode       string
	Region            *string
	RelationshipGoals []profiles.RelationshipGoal
	CreatedAt         time.Time
	PhotoID           *uuid.UUID
}

// presetEntry es una lista ya ordenada. Es inmutable una vez publicada: se
// comparte entre peticiones sin copiarla ni bloquearla.
type presetEntry struct {
	cards   []presetCard
	members map[uuid.UUID]struct{} // profile_id de cards, para descontar exclusiones sin recorrer
	// complete es true si cards contiene TODOS los perfiles activos. Solo
	// entonces los totales y las páginas calculados en memoria son exactos.
	complete bool

	ttl     time.Duration // 0 = TTL por defecto de la caché
	expires time.Time
}

type presetSlot struct {
	mu  sync.Mutex // un solo reconstructor a la vez
	cur atomic.Pointer[presetEntry]
}

type presetCache struct {
	ttl   time.Duration
	slots map[Sort]*presetSlot // fijo tras crearla: se lee sin bloqueo
}

func newPresetCache(ttl time.Duration) *presetCache {
	return &presetCache{
		ttl: ttl,
		slots: map[Sort]*presetSlot{
			SortRecent:  {},
			SortPopular: {},
		},
	}
}

// get devuelve la lista vigente de ese orden, reconstruyéndola con build si
// caducó. Cuando caduca, SOLO UNA petición la reconstruye; las demás siguen
// sirviendo la versión anterior, de modo que no hay estampida de consultas
// pesadas cuando la caché caduca con mucha concurrencia. Si la reconstrucción
// falla y hay una versión anterior, se sigue sirviendo esa.
//
// Devuelve nil (sin error) si ese orden no se cachea.
func (c *presetCache) get(ctx context.Context, sort Sort, build func(context.Context) (*presetEntry, error)) (*presetEntry, error) {
	slot := c.slots[sort]
	if slot == nil {
		return nil, nil
	}

	stale := slot.cur.Load()
	if stale != nil && time.Now().Before(stale.expires) {
		return stale, nil
	}

	if stale != nil {
		if !slot.mu.TryLock() {
			return stale, nil // otra petición ya la está reconstruyendo
		}
	} else {
		slot.mu.Lock() // primera carga: no hay nada que servir mientras tanto
	}
	defer slot.mu.Unlock()

	// Puede que otra petición la haya reconstruido mientras esperábamos.
	if fresh := slot.cur.Load(); fresh != nil && time.Now().Before(fresh.expires) {
		return fresh, nil
	}

	// La carga no debe cancelarse si se cierra la petición que la disparó: la
	// aprovechan todas las demás.
	bctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), presetBuildTimeout)
	defer cancel()

	e, err := build(bctx)
	if err != nil {
		if prev := slot.cur.Load(); prev != nil {
			slog.Warn("search: no se pudo refrescar la caché de listados; se sirve la anterior",
				"sort", string(sort), "error", err)
			return prev, nil
		}
		return nil, err
	}

	ttl := c.ttl
	if e.ttl > 0 {
		ttl = e.ttl
	}
	e.expires = time.Now().Add(ttl)
	slot.cur.Store(e)
	return e, nil
}

// isPlainPreset indica si la búsqueda es un listado sin ningún filtro y con un
// orden cacheable. DeepEqual contra Filters{} hace que cualquier filtro que se
// añada en el futuro (o que venga informado) desactive la caché por sí solo.
func isPlainPreset(p Params) bool {
	if p.Sort != SortRecent && p.Sort != SortPopular {
		return false
	}
	return reflect.DeepEqual(p.Filters, Filters{})
}

// matches indica si una tarjeta cumple las restricciones de quien mira.
func (s ViewerScope) matches(c presetCard, now time.Time) bool {
	if len(s.SeekingGenders) > 0 && !slices.Contains(s.SeekingGenders, c.Gender) {
		return false
	}
	if s.MinAge == nil && s.MaxAge == nil {
		return true
	}
	age := profiles.AgeAt(c.BirthDate, now)
	if s.MinAge != nil && age < *s.MinAge {
		return false
	}
	if s.MaxAge != nil && age > *s.MaxAge {
		return false
	}
	return true
}

// pageOfCards extrae una página de la lista cacheada descontando los perfiles
// excluidos para quien mira (él mismo y bloqueos). Devuelve las tarjetas de la
// página y el total de perfiles visibles para él. Solo es exacta si
// e.complete es true.
func pageOfCards(e *presetEntry, excluded []uuid.UUID, scope ViewerScope, now time.Time, page, pageSize int) (cards []presetCard, total int) {
	skip := make(map[uuid.UUID]struct{}, len(excluded))
	for _, id := range excluded {
		if _, inList := e.members[id]; inList {
			skip[id] = struct{}{}
		}
	}
	if pageSize <= 0 {
		return nil, 0
	}

	offset := (page - 1) * pageSize
	cards = make([]presetCard, 0, pageSize)
	// Una pasada completa: el total exacto depende de las restricciones de
	// quien mira (género/edad), así que hay que contar todas las que cumplen.
	for _, c := range e.cards {
		if _, drop := skip[c.ProfileID]; drop || !scope.matches(c, now) {
			continue
		}
		if total >= offset && len(cards) < pageSize {
			cards = append(cards, c)
		}
		total++
	}
	return cards, total
}
