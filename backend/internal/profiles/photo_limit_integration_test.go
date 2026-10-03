//go:build integration

package profiles

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

// Antes de AddPhoto atómico, N subidas simultáneas pasaban todas el
// count >= Max previo y el perfil acababa con más fotos que el límite.
func TestPostgresRepository_AddPhotoRespectsLimitConcurrently(t *testing.T) {
	pool := testPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	profile := seedTestProfile(t, pool, repo)

	const attempts = 12
	errs := make([]error, attempts)

	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ph := &Photo{
				StorageKey:  fmt.Sprintf("test/%s/%d.jpg", profile.ID, i),
				ContentType: "image/jpeg",
			}
			errs[i] = repo.AddPhoto(ctx, profile.ID, ph, MaxPhotosPerProfile)
		}(i)
	}
	wg.Wait()

	var ok, tooMany int
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, ErrTooManyPhotos):
			tooMany++
		default:
			t.Errorf("error inesperado: %v", err)
		}
	}
	if ok != MaxPhotosPerProfile || tooMany != attempts-MaxPhotosPerProfile {
		t.Errorf("ok=%d tooMany=%d, se esperaba %d y %d", ok, tooMany, MaxPhotosPerProfile, attempts-MaxPhotosPerProfile)
	}

	photos, err := repo.ListPhotos(ctx, profile.ID)
	if err != nil {
		t.Fatalf("ListPhotos: %v", err)
	}
	if len(photos) != MaxPhotosPerProfile {
		t.Fatalf("hay %d fotos, el límite es %d", len(photos), MaxPhotosPerProfile)
	}
	seen := map[int]bool{}
	for _, p := range photos {
		if seen[p.Position] {
			t.Errorf("position %d repetida", p.Position)
		}
		seen[p.Position] = true
	}
}
