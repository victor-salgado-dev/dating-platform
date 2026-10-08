package search

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

func TestViewerScopeCacheCachesCopiesAndInvalidates(t *testing.T) {
	cache := newViewerScopeCache(time.Minute, 10)
	userID := uuid.New()
	minAge := 28
	loads := 0

	load := func() (ViewerScope, error) {
		loads++
		return ViewerScope{SeekingGenders: []profiles.Gender{profiles.GenderFemale}, MinAge: &minAge}, nil
	}

	first, err := cache.getOrLoad(context.Background(), userID, load)
	if err != nil {
		t.Fatal(err)
	}
	*first.MinAge = 18
	first.SeekingGenders[0] = profiles.GenderMale

	second, err := cache.getOrLoad(context.Background(), userID, load)
	if err != nil {
		t.Fatal(err)
	}
	if loads != 1 {
		t.Fatalf("se hicieron %d lecturas; se esperaba 1", loads)
	}
	if second.MinAge == nil || *second.MinAge != 28 || second.SeekingGenders[0] != profiles.GenderFemale {
		t.Fatalf("la caché devolvió datos modificados por el llamador: %+v", second)
	}

	cache.invalidate(userID)
	updated, err := cache.getOrLoad(context.Background(), userID, func() (ViewerScope, error) {
		loads++
		return ViewerScope{SeekingGenders: []profiles.Gender{profiles.GenderMale}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if loads != 2 || updated.SeekingGenders[0] != profiles.GenderMale {
		t.Fatalf("la invalidación no forzó una lectura nueva: loads=%d scope=%+v", loads, updated)
	}
}

func TestViewerScopeCacheInvalidationDuringLoadDoesNotCacheStaleScope(t *testing.T) {
	cache := newViewerScopeCache(time.Minute, 10)
	userID := uuid.New()
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan ViewerScope, 1)
	loads := 0

	go func() {
		scope, _ := cache.getOrLoad(context.Background(), userID, func() (ViewerScope, error) {
			loads++
			if loads == 1 {
				close(started)
				<-release
				return ViewerScope{SeekingGenders: []profiles.Gender{profiles.GenderFemale}}, nil
			}
			return ViewerScope{SeekingGenders: []profiles.Gender{profiles.GenderMale}}, nil
		})
		done <- scope
	}()
	<-started
	cache.invalidate(userID)
	close(release)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("la carga quedó bloqueada")
	}

	updated, err := cache.getOrLoad(context.Background(), userID, func() (ViewerScope, error) {
		loads++
		return ViewerScope{SeekingGenders: []profiles.Gender{profiles.GenderMale}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if loads != 2 || updated.SeekingGenders[0] != profiles.GenderMale {
		t.Fatalf("se guardó un scope anterior a la invalidación: loads=%d scope=%+v", loads, updated)
	}
}
