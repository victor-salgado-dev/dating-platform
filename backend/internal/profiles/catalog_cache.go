package profiles

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// interestCatalogTTL es cuánto se reutiliza el catálogo de intereses antes de
// releerlo. Es dato semi-estático (cambia por migración): una instancia nueva
// lo ve al arrancar, y las ya en marcha en como mucho este tiempo.
const interestCatalogTTL = 10 * time.Minute

// interestCatalogSnapshot es el catálogo ya cargado. SOLO LECTURA: se comparte
// entre peticiones, nadie debe modificar la lista ni sus elementos.
type interestCatalogSnapshot struct {
	list  []InterestDefinition
	byKey map[string]InterestDefinition
}

// interestCatalogCache guarda el catálogo en memoria con caducidad. El valor
// cero está listo para usar.
type interestCatalogCache struct {
	mu      sync.RWMutex
	snap    *interestCatalogSnapshot
	expires time.Time

	now func() time.Time // inyectable en tests; nil = time.Now
}

func (c *interestCatalogCache) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

// get devuelve el catálogo cacheado o lo recarga con load si caducó. Si la
// recarga falla y hay una copia anterior, sirve esa (y lo registra) en vez de
// romper todas las peticiones por un fallo puntual de BD.
func (c *interestCatalogCache) get(ctx context.Context, load func(context.Context) ([]InterestDefinition, error)) (*interestCatalogSnapshot, error) {
	c.mu.RLock()
	snap, exp := c.snap, c.expires
	c.mu.RUnlock()
	if snap != nil && c.clock().Before(exp) {
		return snap, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	// Otra goroutine pudo recargarlo mientras esperábamos el lock.
	if c.snap != nil && c.clock().Before(c.expires) {
		return c.snap, nil
	}

	list, err := load(ctx)
	if err != nil {
		if c.snap != nil {
			slog.Warn("no se pudo refrescar el catálogo de intereses; se sirve la copia anterior", "error", err)
			return c.snap, nil
		}
		return nil, err
	}

	byKey := make(map[string]InterestDefinition, len(list))
	for _, d := range list {
		byKey[d.Key] = d
	}
	c.snap = &interestCatalogSnapshot{list: list, byKey: byKey}
	c.expires = c.clock().Add(interestCatalogTTL)
	return c.snap, nil
}
