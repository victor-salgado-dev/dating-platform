//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"dating-platform/backend/internal/blocking"
	"dating-platform/backend/internal/favorites"
	"dating-platform/backend/internal/likes"
	"dating-platform/backend/internal/profiles"
	"dating-platform/backend/internal/visits"
)

// birthDateISO es la fecha de nacimiento con la que se siembran todos los
// perfiles de prueba. Sirve para comprobar que la edad se calcula a partir
// de la columna correcta: si un SELECT tuviera las columnas desordenadas,
// birth_date llegaría con otro valor y la edad no coincidiría.
const birthDateISO = "1990-01-01"

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL no definida; se salta el test de integración")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("no se pudo crear el pool de conexiones de test: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("no se pudo conectar a la base de datos de test: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

type seededProfile struct {
	userID    uuid.UUID
	profileID uuid.UUID
}

// seedProfile crea un usuario y su perfil, con datos conocidos, para poder
// afirmar después que el listado devuelve exactamente esos valores.
func seedProfile(t *testing.T, pool *pgxpool.Pool, name, gender, country string, region *string) seededProfile {
	t.Helper()
	ctx := context.Background()

	var userID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, status)
		VALUES ($1, 'x', 'active')
		RETURNING id
	`, uuid.NewString()+"@example.test").Scan(&userID); err != nil {
		t.Fatalf("no se pudo crear el usuario de prueba: %v", err)
	}

	var profileID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO profiles (user_id, display_name, birth_date, gender, country_code, region)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, userID, name, birthDateISO, gender, country, region).Scan(&profileID); err != nil {
		t.Fatalf("no se pudo crear el perfil de prueba: %v", err)
	}

	return seededProfile{userID: userID, profileID: profileID}
}

// cleanup borra, en un orden respetuoso con las claves foráneas, todo lo
// que hayan podido dejar los tests. Usa un único UUID por sentencia (sin
// arrays) para no depender de cómo codifica pgx los slices.
func cleanup(t *testing.T, pool *pgxpool.Pool, profileIDs, userIDs []uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	for _, pid := range profileIDs {
		_, _ = pool.Exec(ctx, `DELETE FROM favorites WHERE favorite_profile_id = $1`, pid)
		_, _ = pool.Exec(ctx, `DELETE FROM likes WHERE from_profile_id = $1 OR to_profile_id = $1`, pid)
		_, _ = pool.Exec(ctx, `DELETE FROM matches WHERE profile_one_id = $1 OR profile_two_id = $1`, pid)
		_, _ = pool.Exec(ctx, `DELETE FROM profile_visits WHERE visitor_profile_id = $1 OR visited_profile_id = $1`, pid)
	}
	for _, uid := range userIDs {
		_, _ = pool.Exec(ctx, `DELETE FROM favorites WHERE user_id = $1`, uid)
		_, _ = pool.Exec(ctx, `DELETE FROM blocks WHERE blocker_id = $1 OR blocked_id = $1`, uid)
	}
	for _, pid := range profileIDs {
		_, _ = pool.Exec(ctx, `DELETE FROM profiles WHERE id = $1`, pid)
	}
	for _, uid := range userIDs {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, uid)
	}
}

// assertTarget comprueba que una ficha base trae EXACTAMENTE los valores
// sembrados. Es lo que detecta un desajuste de orden de columnas: si dos
// columnas del mismo tipo se intercambian en el SELECT, el escaneo no
// falla, pero estos campos salen cruzados.
func assertTarget(t *testing.T, base profiles.BaseListItem, wantProfileID uuid.UUID) {
	t.Helper()

	if base.ProfileID != wantProfileID {
		t.Errorf("ProfileID = %v, se esperaba %v", base.ProfileID, wantProfileID)
	}
	if base.DisplayName != "Target" {
		t.Errorf("DisplayName = %q, se esperaba %q", base.DisplayName, "Target")
	}
	if base.Gender != "female" {
		t.Errorf("Gender = %q, se esperaba %q", base.Gender, "female")
	}
	if base.CountryCode != "ES" {
		t.Errorf("CountryCode = %q, se esperaba %q", base.CountryCode, "ES")
	}
	if base.Region == nil || *base.Region != "Madrid" {
		t.Errorf("Region = %v, se esperaba %q", base.Region, "Madrid")
	}
	if base.HasPhoto {
		t.Errorf("HasPhoto = true, se esperaba false")
	}

	wantAge := profiles.AgeAt(mustDate(t, birthDateISO), time.Now())
	if base.Age != wantAge {
		t.Errorf("Age = %d, se esperaba %d", base.Age, wantAge)
	}
}

func mustDate(t *testing.T, iso string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", iso)
	if err != nil {
		t.Fatalf("fecha de prueba inválida %q: %v", iso, err)
	}
	return d
}

func TestListItems_ColumnOrder(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	madrid := "Madrid"

	t.Run("favorites.List mapea las columnas en orden", func(t *testing.T) {
		viewer := seedProfile(t, pool, "Viewer", "male", "ES", nil)
		target := seedProfile(t, pool, "Target", "female", "ES", &madrid)
		t.Cleanup(func() {
			cleanup(t, pool, []uuid.UUID{viewer.profileID, target.profileID}, []uuid.UUID{viewer.userID, target.userID})
		})

		if _, err := pool.Exec(ctx,
			`INSERT INTO favorites (user_id, favorite_profile_id) VALUES ($1, $2)`,
			viewer.userID, target.profileID); err != nil {
			t.Fatalf("no se pudo sembrar el favorito: %v", err)
		}

		res, err := favorites.NewPostgresRepository(pool).List(ctx, viewer.userID, 1, 20)
		if err != nil {
			t.Fatalf("favorites.List devolvió error: %v", err)
		}
		if len(res.Items) != 1 {
			t.Fatalf("se esperaba 1 favorito, hay %d", len(res.Items))
		}
		assertTarget(t, res.Items[0].BaseListItem, target.profileID)
	})

	t.Run("likes.ListSent mapea las columnas en orden", func(t *testing.T) {
		viewer := seedProfile(t, pool, "Viewer", "male", "ES", nil)
		target := seedProfile(t, pool, "Target", "female", "ES", &madrid)
		t.Cleanup(func() {
			cleanup(t, pool, []uuid.UUID{viewer.profileID, target.profileID}, []uuid.UUID{viewer.userID, target.userID})
		})

		if _, err := pool.Exec(ctx,
			`INSERT INTO likes (from_profile_id, to_profile_id) VALUES ($1, $2)`,
			viewer.profileID, target.profileID); err != nil {
			t.Fatalf("no se pudo sembrar el like: %v", err)
		}

		res, err := likes.NewPostgresRepository(pool).ListSent(ctx, viewer.profileID, 1, 20)
		if err != nil {
			t.Fatalf("likes.ListSent devolvió error: %v", err)
		}
		if len(res.Items) != 1 {
			t.Fatalf("se esperaba 1 like enviado, hay %d", len(res.Items))
		}
		assertTarget(t, res.Items[0].BaseListItem, target.profileID)
	})

	t.Run("visits.ListSent mapea las columnas en orden", func(t *testing.T) {
		viewer := seedProfile(t, pool, "Viewer", "male", "ES", nil)
		target := seedProfile(t, pool, "Target", "female", "ES", &madrid)
		t.Cleanup(func() {
			cleanup(t, pool, []uuid.UUID{viewer.profileID, target.profileID}, []uuid.UUID{viewer.userID, target.userID})
		})

		if _, err := pool.Exec(ctx,
			`INSERT INTO profile_visits (visitor_profile_id, visited_profile_id, visited_at) VALUES ($1, $2, now())`,
			viewer.profileID, target.profileID); err != nil {
			t.Fatalf("no se pudo sembrar la visita: %v", err)
		}

		res, err := visits.NewPostgresRepository(pool).ListSent(ctx, viewer.profileID, 1, 20)
		if err != nil {
			t.Fatalf("visits.ListSent devolvió error: %v", err)
		}
		if len(res.Items) != 1 {
			t.Fatalf("se esperaba 1 visita enviada, hay %d", len(res.Items))
		}
		assertTarget(t, res.Items[0].BaseListItem, target.profileID)
	})

	t.Run("blocking.List mapea las columnas en orden", func(t *testing.T) {
		blocker := seedProfile(t, pool, "Viewer", "male", "ES", nil)
		target := seedProfile(t, pool, "Target", "female", "ES", &madrid)
		t.Cleanup(func() {
			cleanup(t, pool, []uuid.UUID{blocker.profileID, target.profileID}, []uuid.UUID{blocker.userID, target.userID})
		})

		if _, err := pool.Exec(ctx,
			`INSERT INTO blocks (blocker_id, blocked_id) VALUES ($1, $2)`,
			blocker.userID, target.userID); err != nil {
			t.Fatalf("no se pudo sembrar el bloqueo: %v", err)
		}

		res, err := blocking.NewPostgresRepository(pool).List(ctx, blocker.userID, 1, 20)
		if err != nil {
			t.Fatalf("blocking.List devolvió error: %v", err)
		}
		if len(res.Items) != 1 {
			t.Fatalf("se esperaba 1 bloqueado, hay %d", len(res.Items))
		}
		assertTarget(t, res.Items[0].BaseListItem, target.profileID)
	})
}
