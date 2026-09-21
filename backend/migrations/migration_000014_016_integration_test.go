//go:build integration

package profiles

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgForeignKeyViolation = "23503"
	pgUniqueViolation     = "23505"
	pgCheckViolation      = "23514"
)

func testTx14(t *testing.T) (pgx.Tx, context.Context) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL no definida; se salta el test de integración")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("no se pudo conectar a la base de datos de test: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(ctx) })

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("no se pudo abrir la transacción: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	return tx, ctx
}

func seedProfile14(t *testing.T, tx pgx.Tx, ctx context.Context) uuid.UUID {
	t.Helper()

	var userID uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, status)
		VALUES ($1, 'x', 'active')
		RETURNING id
	`, uuid.NewString()+"@example.test").Scan(&userID)
	if err != nil {
		t.Fatalf("no se pudo crear el usuario de prueba: %v", err)
	}

	var profileID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO profiles (user_id, display_name, birth_date, gender, country_code)
		VALUES ($1, 'Test User', '1990-01-01', 'female', 'ES')
		RETURNING id
	`, userID).Scan(&profileID)
	if err != nil {
		t.Fatalf("no se pudo crear el perfil de prueba: %v", err)
	}

	return profileID
}

func isPgError14(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

// --- 000014: profile_languages --------------------------------------------

func TestMigration000014_ProfileLanguages(t *testing.T) {
	t.Run("idioma y nivel válidos se aceptan", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		_, err := tx.Exec(ctx, `
			INSERT INTO profile_languages (profile_id, language_code, level)
			VALUES ($1, 'es', 5)
		`, profileID)
		if err != nil {
			t.Errorf("se esperaba que se aceptara: %v", err)
		}
	})

	t.Run("idioma sin nivel se acepta (nivel es opcional)", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		_, err := tx.Exec(ctx, `
			INSERT INTO profile_languages (profile_id, language_code, level)
			VALUES ($1, 'en', NULL)
		`, profileID)
		if err != nil {
			t.Errorf("se esperaba que se aceptara: %v", err)
		}
	})

	t.Run("código de idioma inválido se rechaza", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		_, err := tx.Exec(ctx, `
			INSERT INTO profile_languages (profile_id, language_code, level)
			VALUES ($1, 'klingon', 3)
		`, profileID)
		if !isPgError14(err, pgCheckViolation) {
			t.Errorf("se esperaba una violación de CHECK, se obtuvo: %v", err)
		}
	})

	t.Run("nivel fuera de rango 1-5 se rechaza", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		_, err := tx.Exec(ctx, `
			INSERT INTO profile_languages (profile_id, language_code, level)
			VALUES ($1, 'fr', 9)
		`, profileID)
		if !isPgError14(err, pgCheckViolation) {
			t.Errorf("se esperaba una violación de CHECK, se obtuvo: %v", err)
		}
	})

	t.Run("el mismo idioma dos veces para el mismo perfil se rechaza", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		if _, err := tx.Exec(ctx, `
			INSERT INTO profile_languages (profile_id, language_code, level) VALUES ($1, 'de', 2)
		`, profileID); err != nil {
			t.Fatalf("fallo en el primer insert: %v", err)
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO profile_languages (profile_id, language_code, level) VALUES ($1, 'de', 4)
		`, profileID)
		if !isPgError14(err, pgUniqueViolation) {
			t.Errorf("se esperaba una violación de clave primaria duplicada, se obtuvo: %v", err)
		}
	})

	t.Run("profiles.languages ya no existe", func(t *testing.T) {
		tx, ctx := testTx14(t)
		_, err := tx.Exec(ctx, `SELECT languages FROM profiles LIMIT 1`)
		if err == nil {
			t.Error("se esperaba un error de columna inexistente")
		}
	})
}

// --- 000015: relationship_goals multi-select --------------------------------

func TestMigration000015_RelationshipGoalsMulti(t *testing.T) {
	t.Run("varios objetivos a la vez se aceptan", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		_, err := tx.Exec(ctx, `
			UPDATE profiles SET relationship_goals = ARRAY['casual', 'long_term'] WHERE id = $1
		`, profileID)
		if err != nil {
			t.Errorf("se esperaba que se aceptara: %v", err)
		}
	})

	t.Run("un valor fuera de la lista permitida se rechaza", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		_, err := tx.Exec(ctx, `
			UPDATE profiles SET relationship_goals = ARRAY['casual', 'undecided'] WHERE id = $1
		`, profileID)
		if !isPgError14(err, pgCheckViolation) {
			t.Errorf("se esperaba una violación de CHECK, se obtuvo: %v", err)
		}
	})

	t.Run("array vacío se acepta (distinto de NULL)", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		_, err := tx.Exec(ctx, `
			UPDATE profiles SET relationship_goals = ARRAY[]::TEXT[] WHERE id = $1
		`, profileID)
		if err != nil {
			t.Errorf("se esperaba que se aceptara: %v", err)
		}
	})

	t.Run("profiles.relationship_goal (singular) ya no existe", func(t *testing.T) {
		tx, ctx := testTx14(t)
		_, err := tx.Exec(ctx, `SELECT relationship_goal FROM profiles LIMIT 1`)
		if err == nil {
			t.Error("se esperaba un error de columna inexistente")
		}
	})
}

// --- 000016: interests / profile_interests ----------------------------------

func TestMigration000016_InterestsCatalog(t *testing.T) {
	tx, ctx := testTx14(t)

	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM interests`).Scan(&count); err != nil {
		t.Fatalf("no se pudo contar interests: %v", err)
	}
	if count != 291 {
		t.Errorf("se esperaban 291 intereses en el catálogo, hay %d", count)
	}

	t.Run("los ítems amplios tienen has_level=true", func(t *testing.T) {
		var hasLevel bool
		if err := tx.QueryRow(ctx, `SELECT has_level FROM interests WHERE key = 'escalada'`).Scan(&hasLevel); err != nil {
			t.Fatalf("no se encontró el interés 'escalada': %v", err)
		}
		if !hasLevel {
			t.Error("'escalada' (grupo con intensidad) debería tener has_level=true")
		}
	})

	t.Run("los ítems concretos tienen has_level=false", func(t *testing.T) {
		var hasLevel bool
		if err := tx.QueryRow(ctx, `SELECT has_level FROM interests WHERE key = 'guitarra'`).Scan(&hasLevel); err != nil {
			t.Fatalf("no se encontró el interés 'guitarra': %v", err)
		}
		if hasLevel {
			t.Error("'guitarra' (grupo concreto) debería tener has_level=false")
		}
	})

	t.Run("los duplicados desambiguados por sufijo existen como filas distintas", func(t *testing.T) {
		var filmLabel, bookLabel string
		if err := tx.QueryRow(ctx, `SELECT label FROM interests WHERE key = 'romance_film'`).Scan(&filmLabel); err != nil {
			t.Fatalf("no se encontró 'romance_film': %v", err)
		}
		if err := tx.QueryRow(ctx, `SELECT label FROM interests WHERE key = 'romance_book'`).Scan(&bookLabel); err != nil {
			t.Fatalf("no se encontró 'romance_book': %v", err)
		}
	})

	t.Run("profiles.interests (el array libre) ya no existe", func(t *testing.T) {
		_, err := tx.Exec(ctx, `SELECT interests FROM profiles LIMIT 1`)
		if err == nil {
			t.Error("se esperaba un error de columna inexistente")
		}
	})

	t.Run("hobby_definitions/profile_hobbies (000013) ya no existen", func(t *testing.T) {
		_, err := tx.Exec(ctx, `SELECT 1 FROM hobby_definitions LIMIT 1`)
		if err == nil {
			t.Error("se esperaba un error de tabla inexistente (hobby_definitions)")
		}
	})
}

func TestMigration000016_ProfileInterests(t *testing.T) {
	t.Run("interés con nivel válido se acepta", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		_, err := tx.Exec(ctx, `
			INSERT INTO profile_interests (profile_id, interest_key, level)
			VALUES ($1, 'cocina', 4)
		`, profileID)
		if err != nil {
			t.Errorf("se esperaba que se aceptara: %v", err)
		}
	})

	t.Run("interés concreto sin nivel se acepta", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		_, err := tx.Exec(ctx, `
			INSERT INTO profile_interests (profile_id, interest_key, level)
			VALUES ($1, 'guitarra', NULL)
		`, profileID)
		if err != nil {
			t.Errorf("se esperaba que se aceptara: %v", err)
		}
	})

	t.Run("nivel fuera de rango 1-5 se rechaza", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		_, err := tx.Exec(ctx, `
			INSERT INTO profile_interests (profile_id, interest_key, level)
			VALUES ($1, 'lectura', 7)
		`, profileID)
		if !isPgError14(err, pgCheckViolation) {
			t.Errorf("se esperaba una violación de CHECK, se obtuvo: %v", err)
		}
	})

	t.Run("interest_key inexistente en el catálogo se rechaza", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		_, err := tx.Exec(ctx, `
			INSERT INTO profile_interests (profile_id, interest_key, level)
			VALUES ($1, 'no_existe_este_interes', 3)
		`, profileID)
		if !isPgError14(err, pgForeignKeyViolation) {
			t.Errorf("se esperaba una violación de FOREIGN KEY, se obtuvo: %v", err)
		}
	})

	t.Run("el mismo interés dos veces para el mismo perfil se rechaza", func(t *testing.T) {
		tx, ctx := testTx14(t)
		profileID := seedProfile14(t, tx, ctx)
		if _, err := tx.Exec(ctx, `
			INSERT INTO profile_interests (profile_id, interest_key, level) VALUES ($1, 'senderismo', NULL)
		`, profileID); err != nil {
			t.Fatalf("fallo en el primer insert: %v", err)
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO profile_interests (profile_id, interest_key, level) VALUES ($1, 'senderismo', NULL)
		`, profileID)
		if !isPgError14(err, pgUniqueViolation) {
			t.Errorf("se esperaba una violación de clave primaria duplicada, se obtuvo: %v", err)
		}
	})
}