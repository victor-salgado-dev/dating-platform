package search

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Caché en memoria del TOTAL de resultados de una búsqueda con filtros.
//
// Contar cuántos perfiles cumplen los filtros obliga a recorrerlos todos (con
// 50.000 perfiles es lo más caro de una búsqueda), mientras que la página en sí
// se resuelve parando en cuanto se juntan PageSize filas. Por eso el total se
// calcula aparte y se recuerda unos segundos: paginar por la misma búsqueda no
// repite el recuento.
//
// El total puede ir desfasado hasta countCacheTTL (un perfil nuevo, un cambio
// de filtros de otra cuenta...). El recuento no depende de quien busca (ver
// buildSearchQuery), así que la misma entrada la comparten todos los usuarios
// con los mismos filtros.
const (
	countCacheTTL = 45 * time.Second
	countCacheMax = 20_000 // acota la memoria: unos pocos MB como mucho
)

type countEntry struct {
	total   int
	expires time.Time
}

type countCache struct {
	mu  sync.Mutex
	ttl time.Duration
	max int
	m   map[string]countEntry
}

func newCountCache(ttl time.Duration, max int) *countCache {
	return &countCache{ttl: ttl, max: max, m: make(map[string]countEntry)}
}

// countKey resume la consulta de recuento y sus argumentos en una clave corta.
func countKey(q builtQuery) string {
	sum := sha256.Sum256([]byte(fmt.Sprint(q.CountSQL, q.CountArgs)))
	return hex.EncodeToString(sum[:16])
}

func (c *countCache) get(key string) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[key]
	if !ok {
		return 0, false
	}
	if time.Now().After(e.expires) {
		delete(c.m, key)
		return 0, false
	}
	return e.total, true
}

func (c *countCache) set(key string, total int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.m) >= c.max {
		now := time.Now()
		for k, e := range c.m {
			if now.After(e.expires) {
				delete(c.m, k)
			}
		}
		if len(c.m) >= c.max {
			c.m = make(map[string]countEntry) // todo vigente: se reinicia en bloque
		}
	}
	c.m[key] = countEntry{total: total, expires: time.Now().Add(c.ttl)}
}
