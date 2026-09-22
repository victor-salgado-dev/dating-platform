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

// --- infraestructura del test ---------------------------------------

func testTx(t *testing.T) pgx.Tx {
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

	return tx
}

func seedProfile(t *testing.T, tx pgx.Tx) uuid.UUID {
	t.Helper()
	ctx := context.Background()

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

func isPgError(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

const (
	pgCheckViolationCode  = "23514"
	pgForeignKeyCode      = "23503"
	pgUniqueViolationCode = "23505"
)

// --- 1. "Über mich": nuevos campos en profiles -----------------------

func TestMigration000013_LifestyleFields(t *testing.T) {
	ctx := context.Background()

	t.Run("valores válidos se aceptan", func(t *testing.T) {
		tx := testTx(t)
		profileID := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `
			UPDATE profiles SET
				sports = ARRAY['fitness', 'cycling', 'hiking'],
				likes_pets = 'yes',
				pets_owned = ARRAY['cat'],
				favorite_season = 'summer',
				ideal_vacation_style = ARRAY['small_charming_hotel', 'countryside_house'],
				vacation_activities = ARRAY['mix_relaxation_activities'],
				future_vision = ARRAY['balance_family_career'],
				profile_quote = 'Me encanta viajar.',
				dream_wish = 'Visitar todos los rincones del mundo.'
			WHERE id = $1
		`, profileID)
		if err != nil {
			t.Errorf("se esperaba que los valores válidos se aceptaran: %v", err)
		}
	})

	t.Run("valor fuera de la lista permitida en sports se rechaza", func(t *testing.T) {
		tx := testTx(t)
		profileID := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `UPDATE profiles SET sports = ARRAY['esports'] WHERE id = $1`, profileID)
		if !isPgError(err, pgCheckViolationCode) {
			t.Errorf("se esperaba una violación de CHECK, se obtuvo: %v", err)
		}
	})

	t.Run("valor fuera de la lista permitida en likes_pets se rechaza", func(t *testing.T) {
		tx := testTx(t)
		profileID := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `UPDATE profiles SET likes_pets = 'maybe' WHERE id = $1`, profileID)
		if !isPgError(err, pgCheckViolationCode) {
			t.Errorf("se esperaba una violación de CHECK, se obtuvo: %v", err)
		}
	})

	t.Run("profile_quote demasiado larga se rechaza", func(t *testing.T) {
		tx := testTx(t)
		profileID := seedProfile(t, tx)

		tooLong := make([]byte, 501)
		for i := range tooLong {
			tooLong[i] = 'a'
		}
		_, err := tx.Exec(ctx, `UPDATE profiles SET profile_quote = $2 WHERE id = $1`, profileID, string(tooLong))
		if !isPgError(err, pgCheckViolationCode) {
			t.Errorf("se esperaba una violación de CHECK por longitud, se obtuvo: %v", err)
		}
	})
}

// --- 2. Personalidad ----------------------------------------------------

func TestMigration000013_PersonalityCatalog(t *testing.T) {
	tx := testTx(t)
	ctx := context.Background()

	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM personality_statements`).Scan(&count); err != nil {
		t.Fatalf("no se pudo contar personality_statements: %v", err)
	}
	if count != 20 {
		t.Errorf("se esperaban 20 afirmaciones, hay %d", count)
	}

	rows, err := tx.Query(ctx, `
		SELECT trait_key, COUNT(*) FROM personality_statements GROUP BY trait_key
	`)
	if err != nil {
		t.Fatalf("no se pudo agrupar por trait_key: %v", err)
	}
	defer rows.Close()

	traitCounts := map[string]int{}
	for rows.Next() {
		var trait string
		var n int
		if err := rows.Scan(&trait, &n); err != nil {
			t.Fatalf("error leyendo fila: %v", err)
		}
		traitCounts[trait] = n
	}
	for _, trait := range []string{"extraversion", "emotional_stability", "conscientiousness", "agreeableness", "openness"} {
		if traitCounts[trait] != 4 {
			t.Errorf("se esperaban 4 afirmaciones para %q, hay %d", trait, traitCounts[trait])
		}
	}

	t.Run("trait_key inválido en el catálogo se rechaza", func(t *testing.T) {
		txSub := testTx(t)
		_, err := txSub.Exec(ctx, `
			INSERT INTO personality_statements (key, trait_key, label, sort_order)
			VALUES ('bogus_statement', 'not_a_real_trait', 'x', 1)
		`)
		if !isPgError(err, pgCheckViolationCode) {
			t.Errorf("se esperaba una violación de CHECK, se obtuvo: %v", err)
		}
	})
}

func TestMigration000013_ProfilePersonalityAnswers(t *testing.T) {
	ctx := context.Background()

	t.Run("respuesta válida se acepta", func(t *testing.T) {
		tx := testTx(t)
		profileID := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `
			INSERT INTO profile_personality_answers (profile_id, statement_key, score)
			VALUES ($1, 'extra_funny_laughs', 4)
		`, profileID)
		if err != nil {
			t.Errorf("se esperaba que la respuesta válida se aceptara: %v", err)
		}
	})

	t.Run("score fuera de rango 1-5 se rechaza", func(t *testing.T) {
		tx := testTx(t)
		profileID := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `
			INSERT INTO profile_personality_answers (profile_id, statement_key, score)
			VALUES ($1, 'extra_reserved_calm', 0)
		`, profileID)
		if !isPgError(err, pgCheckViolationCode) {
			t.Errorf("se esperaba una violación de CHECK por rango, se obtuvo: %v", err)
		}
	})

	t.Run("statement_key inexistente se rechaza", func(t *testing.T) {
		tx := testTx(t)
		profileID := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `
			INSERT INTO profile_personality_answers (profile_id, statement_key, score)
			VALUES ($1, 'not_a_real_statement', 3)
		`, profileID)
		if !isPgError(err, pgForeignKeyCode) {
			t.Errorf("se esperaba una violación de FOREIGN KEY, se obtuvo: %v", err)
		}
	})

	t.Run("la vista de agregados calcula bien la media por rasgo", func(t *testing.T) {
		tx := testTx(t)
		profileID := seedProfile(t, tx)

		for _, s := range []struct {
			key   string
			score int
		}{
			{"extra_funny_laughs", 4},
			{"extra_center_of_party", 2},
			{"extra_enjoys_alone_time", 4},
		} {
			if _, err := tx.Exec(ctx, `
				INSERT INTO profile_personality_answers (profile_id, statement_key, score)
				VALUES ($1, $2, $3)
			`, profileID, s.key, s.score); err != nil {
				t.Fatalf("no se pudo insertar respuesta semilla: %v", err)
			}
		}

		var avg float64
		var answered int
		err := tx.QueryRow(ctx, `
			SELECT avg_score, answered_count FROM profile_personality_trait_scores
			WHERE profile_id = $1 AND trait_key = 'extraversion'
		`, profileID).Scan(&avg, &answered)
		if err != nil {
			t.Fatalf("no se pudo leer profile_personality_trait_scores: %v", err)
		}
		if answered != 3 {
			t.Errorf("se esperaban 3 respuestas contabilizadas, hay %d", answered)
		}
		if avg < 3.32 || avg > 3.34 {
			t.Errorf("se esperaba una media ≈3.33, se obtuvo %v", avg)
		}
	})
}

// --- 3. Preferencias de pareja -------------------------------------------

func TestMigration000013_PartnerPreferences(t *testing.T) {
	ctx := context.Background()

	t.Run("preferencias válidas se aceptan", func(t *testing.T) {
		tx := testTx(t)
		profileID := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `
			INSERT INTO profile_partner_preferences (
				profile_id, age_min, age_max, height_min, height_max,
				desired_traits, partner_may_have_children, first_meeting_preference,
				importance_shared_thoughts, importance_intimacy
			) VALUES (
				$1, 38, 58, 175, 200,
				ARRAY['humorous', 'kind_hearted', 'faithful'], 'yes', 'doesnt_matter',
				5, 4
			)
		`, profileID)
		if err != nil {
			t.Errorf("se esperaba que las preferencias válidas se aceptaran: %v", err)
		}
	})

	t.Run("age_min mayor que age_max se rechaza", func(t *testing.T) {
		tx := testTx(t)
		otherProfile := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `
			INSERT INTO profile_partner_preferences (profile_id, age_min, age_max)
			VALUES ($1, 50, 30)
		`, otherProfile)
		if !isPgError(err, pgCheckViolationCode) {
			t.Errorf("se esperaba una violación de CHECK por rango de edad, se obtuvo: %v", err)
		}
	})

	t.Run("height_min mayor que height_max se rechaza", func(t *testing.T) {
		tx := testTx(t)
		otherProfile := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `
			INSERT INTO profile_partner_preferences (profile_id, height_min, height_max)
			VALUES ($1, 190, 160)
		`, otherProfile)
		if !isPgError(err, pgCheckViolationCode) {
			t.Errorf("se esperaba una violación de CHECK por rango de altura, se obtuvo: %v", err)
		}
	})

	t.Run("importancia fuera de rango 1-5 se rechaza", func(t *testing.T) {
		tx := testTx(t)
		otherProfile := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `
			INSERT INTO profile_partner_preferences (profile_id, importance_fun)
			VALUES ($1, 9)
		`, otherProfile)
		if !isPgError(err, pgCheckViolationCode) {
			t.Errorf("se esperaba una violación de CHECK por rango, se obtuvo: %v", err)
		}
	})

	t.Run("rasgo deseado fuera de la lista permitida se rechaza", func(t *testing.T) {
		tx := testTx(t)
		otherProfile := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `
			INSERT INTO profile_partner_preferences (profile_id, desired_traits)
			VALUES ($1, ARRAY['telepathic'])
		`, otherProfile)
		if !isPgError(err, pgCheckViolationCode) {
			t.Errorf("se esperaba una violación de CHECK, se obtuvo: %v", err)
		}
	})

	t.Run("una segunda fila de preferencias para el mismo perfil se rechaza (1:1)", func(t *testing.T) {
		tx := testTx(t)
		profileID := seedProfile(t, tx)

		_, err := tx.Exec(ctx, `
			INSERT INTO profile_partner_preferences (profile_id) VALUES ($1)
		`, profileID)
		if err != nil {
			t.Fatalf("fallo al insertar primera fila: %v", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO profile_partner_preferences (profile_id) VALUES ($1)
		`, profileID)
		if !isPgError(err, pgUniqueViolationCode) {
			t.Errorf("se esperaba una violación de clave primaria duplicada, se obtuvo: %v", err)
		}
	})
}