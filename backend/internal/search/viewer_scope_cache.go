package search

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	viewerScopeCacheTTL = time.Minute
	viewerScopeCacheMax = 50_000
)

type viewerScopeEntry struct {
	scope   ViewerScope
	expires time.Time
}

type viewerScopeCall struct {
	done       chan struct{}
	scope      ViewerScope
	err        error
	generation uint64
}

// viewerScopeCache keeps each user's tiny search scope in memory. Profile
// updates invalidate the entry; the TTL also bounds staleness for writes made
// outside this process.
type viewerScopeCache struct {
	mu         sync.Mutex
	ttl        time.Duration
	max        int
	entries    map[uuid.UUID]viewerScopeEntry
	inflight   map[uuid.UUID]*viewerScopeCall
	generation map[uuid.UUID]uint64
}

func newViewerScopeCache(ttl time.Duration, max int) *viewerScopeCache {
	return &viewerScopeCache{
		ttl:        ttl,
		max:        max,
		entries:    make(map[uuid.UUID]viewerScopeEntry),
		inflight:   make(map[uuid.UUID]*viewerScopeCall),
		generation: make(map[uuid.UUID]uint64),
	}
}

func (c *viewerScopeCache) getOrLoad(ctx context.Context, userID uuid.UUID, load func() (ViewerScope, error)) (ViewerScope, error) {
	for {
		c.mu.Lock()
		if entry, ok := c.entries[userID]; ok {
			if time.Now().Before(entry.expires) {
				c.mu.Unlock()
				return cloneViewerScope(entry.scope), nil
			}
			delete(c.entries, userID)
		}
		if call, ok := c.inflight[userID]; ok {
			c.mu.Unlock()
			select {
			case <-call.done:
				c.mu.Lock()
				stale := c.generation[userID] != call.generation
				c.mu.Unlock()
				if stale {
					continue
				}
				if ctx.Err() == nil && (errors.Is(call.err, context.Canceled) || errors.Is(call.err, context.DeadlineExceeded)) {
					continue
				}
				return cloneViewerScope(call.scope), call.err
			case <-ctx.Done():
				return ViewerScope{}, ctx.Err()
			}
		}
		call := &viewerScopeCall{done: make(chan struct{})}
		call.generation = c.generation[userID]
		c.inflight[userID] = call
		c.mu.Unlock()

		call.scope, call.err = load()

		c.mu.Lock()
		stale := c.generation[userID] != call.generation
		if call.err == nil && !stale {
			c.setLocked(userID, call.scope)
		}
		delete(c.inflight, userID)
		close(call.done)
		c.mu.Unlock()
		if stale {
			continue
		}
		return cloneViewerScope(call.scope), call.err
	}
}

func (c *viewerScopeCache) invalidate(userID uuid.UUID) {
	c.mu.Lock()
	delete(c.entries, userID)
	c.generation[userID]++
	c.mu.Unlock()
}

func (c *viewerScopeCache) setLocked(userID uuid.UUID, scope ViewerScope) {
	if len(c.entries) >= c.max {
		now := time.Now()
		for id, entry := range c.entries {
			if now.After(entry.expires) {
				delete(c.entries, id)
			}
		}
		if len(c.entries) >= c.max {
			c.entries = make(map[uuid.UUID]viewerScopeEntry)
		}
	}
	c.entries[userID] = viewerScopeEntry{scope: cloneViewerScope(scope), expires: time.Now().Add(c.ttl)}
}

func cloneViewerScope(scope ViewerScope) ViewerScope {
	copy := ViewerScope{}
	if len(scope.SeekingGenders) > 0 {
		copy.SeekingGenders = append(scope.SeekingGenders[:0:0], scope.SeekingGenders...)
	}
	if scope.MinAge != nil {
		v := *scope.MinAge
		copy.MinAge = &v
	}
	if scope.MaxAge != nil {
		v := *scope.MaxAge
		copy.MaxAge = &v
	}
	return copy
}
