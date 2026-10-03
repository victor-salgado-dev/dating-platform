//go:build integration

package profiles

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// La regla de visibilidad vive en UNA constante (visiblePredicate) usada por
// GetPublicByID, IsVisible, GetPublicPhoto e IDResolver.ResolveTarget. Este
// test comprueba que las cuatro dan el mismo veredicto en cada escenario.
func TestVisibilityPredicate_AllEntryPointsAgree(t *testing.T) {
	pool := testPool(t)
	repo := NewPostgresRepository(pool)
	resolver := NewIDResolver(pool)
	ctx := context.Background()

	target := seedTestProfile(t, pool, repo) // el perfil que se mira
	viewer := seedTestProfile(t, pool, repo) // quien mira

	photo := &Photo{StorageKey: "test/" + target.ID.String() + "/v.jpg", ContentType: "image/jpeg"}
	if err := repo.AddPhoto(ctx, target.ID, photo, MaxPhotosPerProfile); err != nil {
		t.Fatalf("AddPhoto: %v", err)
	}

	// check devuelve true si las cuatro entradas dicen "visible" y falla el
	// test si alguna discrepa de las demás.
	check := func(t *testing.T) bool {
		t.Helper()
		_, e1 := repo.GetPublicByID(ctx, target.ID, viewer.UserID)
		e2 := repo.IsVisible(ctx, target.ID, viewer.UserID)
		_, e3 := repo.GetPublicPhoto(ctx, target.ID, viewer.UserID, photo.ID)
		_, e4 := resolver.ResolveTarget(ctx, viewer.UserID, target.ID)

		vis := []bool{e1 == nil, e2 == nil, e3 == nil, e4 == nil}
		for i, v := range vis {
			if v != vis[0] {
				t.Fatalf("las entradas discrepan: %v (errores: %v | %v | %v | %v)", vis, e1, e2, e3, e4)
			}
			if !v {
				var want error = ErrNotFound
				if i == 2 {
					want = ErrPhotoNotFound
				}
				got := []error{e1, e2, e3, e4}[i]
				if !errors.Is(got, want) {
					t.Errorf("entrada %d: error %v, se esperaba %v", i, got, want)
				}
			}
		}
		return vis[0]
	}

	block := func(t *testing.T, blocker, blocked uuid.UUID) {
		t.Helper()
		if _, err := pool.Exec(ctx, `INSERT INTO blocks (blocker_id, blocked_id) VALUES ($1, $2)`, blocker, blocked); err != nil {
			t.Fatalf("crear bloqueo: %v", err)
		}
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM blocks WHERE blocker_id = $1 AND blocked_id = $2`, blocker, blocked)
		})
	}

	t.Run("visible por defecto", func(t *testing.T) {
		if !check(t) {
			t.Error("se esperaba visible")
		}
	})

	t.Run("el visitante bloquea al perfil", func(t *testing.T) {
		block(t, viewer.UserID, target.UserID)
		if check(t) {
			t.Error("se esperaba invisible")
		}
	})

	t.Run("el perfil bloquea al visitante", func(t *testing.T) {
		block(t, target.UserID, viewer.UserID)
		if check(t) {
			t.Error("se esperaba invisible")
		}
	})

	t.Run("cuenta eliminada", func(t *testing.T) {
		if _, err := pool.Exec(ctx, `UPDATE users SET deleted_at = now() WHERE id = $1`, target.UserID); err != nil {
			t.Fatalf("marcar cuenta eliminada: %v", err)
		}
		if check(t) {
			t.Error("se esperaba invisible")
		}
	})
}
