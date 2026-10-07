package search

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

type countCall struct {
	done  chan struct{}
	total int
	err   error
}

type countCache struct {
	mu       sync.Mutex
	ttl      time.Duration
	max      int
	m        map[string]countEntry
	inflight map[string]*countCall
}

func newCountCache(ttl time.Duration, max int) *countCache {
	return &countCache{
		ttl:      ttl,
		max:      max,
		m:        make(map[string]countEntry),
		inflight: make(map[string]*countCall),
	}
}

// countKey resume la consulta de recuento y sus argumentos en una clave corta.
func countKey(q builtQuery) string {
	sum := sha256.Sum256([]byte(fmt.Sprint(q.CountSQL, q.CountArgs)))
	return hex.EncodeToString(sum[:16])
}

// getOrLoad stores one calculation per key at a time. Concurrent misses wait
// for the same database query instead of stampeding the database.
func (c *countCache) getOrLoad(ctx context.Context, key string, load func() (int, error)) (int, error) {
	for {
		c.mu.Lock()
		e, ok := c.m[key]
		if ok && time.Now().Before(e.expires) {
			c.mu.Unlock()
			return e.total, nil
		}
		if ok {
			delete(c.m, key)
		}
		if call, ok := c.inflight[key]; ok {
			c.mu.Unlock()
			select {
			case <-call.done:
				if ctx.Err() == nil && (errors.Is(call.err, context.Canceled) || errors.Is(call.err, context.DeadlineExceeded)) {
					continue
				}
				return call.total, call.err
			case <-ctx.Done():
				return 0, ctx.Err()
			}
		}
		call := &countCall{done: make(chan struct{})}
		c.inflight[key] = call
		c.mu.Unlock()

		call.total, call.err = load()

		c.mu.Lock()
		if call.err == nil {
			c.setLocked(key, call.total)
		}
		delete(c.inflight, key)
		close(call.done)
		c.mu.Unlock()
		return call.total, call.err
	}
}

func (c *countCache) setLocked(key string, total int) {
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
