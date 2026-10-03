//go:build integration

// Test de integración: ejercita el código Go de PostgresRepository tal
// cual lo usará el Service, contra una base real con las migraciones
// aplicadas hasta la 000016 incluida.
//
//	migrate -database "$TEST_DATABASE_URL" -path ./migrations up
//	TEST_DATABASE_URL="postgres://..." go test -tags=integration ./internal/profiles/... -v
package profiles

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL no definida; se salta el test de integración")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("no se pudo crear el pool de conexiones: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

// seedTestProfile crea un usuario y un perfil mínimos válidos usando el
// propio PostgresRepository, y registra su borrado al final del test
// (el ON DELETE CASCADE se lleva perfil, idiomas, intereses, respuestas
// de personalidad y preferencias de pareja con él).
func seedTestProfile(t *testing.T, pool *pgxpool.Pool, repo *PostgresRepository) *Profile {
	t.Helper()
	ctx := context.Background()

	var userID uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, status)
		VALUES ($1, 'x', 'active')
		RETURNING id
	`, uuid.NewString()+"@example.test").Scan(&userID)
	if err != nil {
		t.Fatalf("no se pudo crear el usuario de prueba: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	p := &Profile{
		UserID:      userID,
		DisplayName: "Test User",
		BirthDate:   date(1990, 1, 1),
		Gender:      GenderFemale,
		CountryCode: "ES",
	}
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("no se pudo crear el perfil de prueba: %v", err)
	}

	return p
}

// --- Idiomas ----------------------------------------------------------

func TestPostgresRepository_ProfileLanguages(t *testing.T) {
	pool := testPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	profile := seedTestProfile(t, pool, repo)

	level := 5
	pl, err := repo.UpsertProfileLanguage(ctx, profile.ID, "es", &level)
	if err != nil {
		t.Fatalf("UpsertProfileLanguage: %v", err)
	}
	if pl.Level == nil || *pl.Level != 5 {
		t.Errorf("resultado inesperado: %+v", pl)
	}

	list, err := repo.ListProfileLanguages(ctx, profile.ID)
	if err != nil {
		t.Fatalf("ListProfileLanguages: %v", err)
	}
	if len(list) != 1 || list[0].LanguageCode != "es" {
		t.Fatalf("se esperaba 1 idioma 'es', se obtuvo: %+v", list)
	}

	// Upsert: bajar el nivel de un idioma ya indicado no crea fila nueva.
	lower := 2
	if _, err := repo.UpsertProfileLanguage(ctx, profile.ID, "es", &lower); err != nil {
		t.Fatalf("UpsertProfileLanguage (update): %v", err)
	}
	list, err = repo.ListProfileLanguages(ctx, profile.ID)
	if err != nil {
		t.Fatalf("ListProfileLanguages tras actualizar: %v", err)
	}
	if len(list) != 1 || list[0].Level == nil || *list[0].Level != 2 {
		t.Fatalf("se esperaba nivel actualizado a 2, se obtuvo: %+v", list)
	}

	if err := repo.DeleteProfileLanguage(ctx, profile.ID, "es"); err != nil {
		t.Fatalf("DeleteProfileLanguage: %v", err)
	}
	list, err = repo.ListProfileLanguages(ctx, profile.ID)
	if err != nil {
		t.Fatalf("ListProfileLanguages tras borrar: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("se esperaba la lista vacía tras borrar, hay %d elementos", len(list))
	}

	// Código de idioma inválido: lo rechaza el CHECK, traducido a invalidField.
	if _, err := repo.UpsertProfileLanguage(ctx, profile.ID, "klingon", nil); err == nil {
		t.Error("se esperaba que un código de idioma inválido se rechazara")
	}
}

// --- Catálogo de intereses -----------------------------------------------

func TestPostgresRepository_InterestCatalog(t *testing.T) {
	pool := testPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	defs, err := repo.ListInterestDefinitions(ctx)
	if err != nil {
		t.Fatalf("ListInterestDefinitions: %v", err)
	}
	if len(defs) != 291 {
		t.Errorf("se esperaban 291 intereses en el catálogo, hay %d", len(defs))
	}

	d, err := repo.GetInterestDefinition(ctx, "cocina")
	if err != nil {
		t.Fatalf("GetInterestDefinition('cocina'): %v", err)
	}
	if !d.HasLevel {
		t.Error("'cocina' debería tener has_level=true")
	}

	d, err = repo.GetInterestDefinition(ctx, "guitarra")
	if err != nil {
		t.Fatalf("GetInterestDefinition('guitarra'): %v", err)
	}
	if d.HasLevel {
		t.Error("'guitarra' debería tener has_level=false")
	}

	if _, err := repo.GetInterestDefinition(ctx, "no_existe_esto"); !errors.Is(err, ErrInterestNotFound) {
		t.Errorf("se esperaba ErrInterestNotFound, se obtuvo: %v", err)
	}
}

// --- Intereses del perfil ---------------------------------------------------

func TestPostgresRepository_ProfileInterests(t *testing.T) {
	pool := testPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	profile := seedTestProfile(t, pool, repo)

	t.Run("interés con nivel se guarda y se lee", func(t *testing.T) {
		level := 4
		pi, err := repo.UpsertProfileInterest(ctx, profile.ID, "cocina", &level)
		if err != nil {
			t.Fatalf("UpsertProfileInterest: %v", err)
		}
		if pi.Level == nil || *pi.Level != 4 {
			t.Errorf("resultado inesperado: %+v", pi)
		}
	})

	t.Run("interés concreto sin nivel se guarda y se lee", func(t *testing.T) {
		pi, err := repo.UpsertProfileInterest(ctx, profile.ID, "guitarra", nil)
		if err != nil {
			t.Fatalf("UpsertProfileInterest: %v", err)
		}
		if pi.Level != nil {
			t.Errorf("se esperaba level=nil, se obtuvo %v", *pi.Level)
		}
	})

	t.Run("ListProfileInterests devuelve ambos", func(t *testing.T) {
		list, err := repo.ListProfileInterests(ctx, profile.ID)
		if err != nil {
			t.Fatalf("ListProfileInterests: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("se esperaban 2 intereses, hay %d", len(list))
		}
	})

	t.Run("nivel fuera de rango 1-5 se rechaza", func(t *testing.T) {
		bad := 9
		if _, err := repo.UpsertProfileInterest(ctx, profile.ID, "lectura", &bad); err == nil {
			t.Error("se esperaba que se rechazara")
		}
	})

	t.Run("interest_key inexistente en el catálogo se rechaza (FK)", func(t *testing.T) {
		if _, err := repo.UpsertProfileInterest(ctx, profile.ID, "no_existe_esto", nil); err == nil {
			t.Error("se esperaba que se rechazara")
		}
	})

	t.Run("DeleteProfileInterest vuelve a 'no seleccionado'", func(t *testing.T) {
		if err := repo.DeleteProfileInterest(ctx, profile.ID, "guitarra"); err != nil {
			t.Fatalf("DeleteProfileInterest: %v", err)
		}
		list, err := repo.ListProfileInterests(ctx, profile.ID)
		if err != nil {
			t.Fatalf("ListProfileInterests: %v", err)
		}
		for _, pi := range list {
			if pi.InterestKey == "guitarra" {
				t.Error("'guitarra' debería haber desaparecido de la lista")
			}
		}
	})
}

// --- Personalidad -----------------------------------------------------

func TestPostgresRepository_PersonalityAnswersAndScores(t *testing.T) {
	pool := testPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	profile := seedTestProfile(t, pool, repo)

	stmts, err := repo.ListPersonalityStatements(ctx)
	if err != nil {
		t.Fatalf("ListPersonalityStatements: %v", err)
	}
	if len(stmts) != 20 {
		t.Errorf("se esperaban 20 afirmaciones, se obtuvieron %d", len(stmts))
	}

	answers := map[string]int{
		"extra_reserved_calm":     2,
		"extra_funny_laughs":      4,
		"extra_center_of_party":   3,
		"extra_enjoys_alone_time": 5,
	}
	for key, score := range answers {
		if _, err := repo.UpsertPersonalityAnswer(ctx, profile.ID, key, score); err != nil {
			t.Fatalf("UpsertPersonalityAnswer(%s): %v", key, err)
		}
	}

	list, err := repo.ListPersonalityAnswers(ctx, profile.ID)
	if err != nil {
		t.Fatalf("ListPersonalityAnswers: %v", err)
	}
	if len(list) != 4 {
		t.Fatalf("se esperaban 4 respuestas, se obtuvieron %d", len(list))
	}

	scores, err := repo.GetPersonalityTraitScores(ctx, profile.ID)
	if err != nil {
		t.Fatalf("GetPersonalityTraitScores: %v", err)
	}
	var extraversion *PersonalityTraitScore
	for i := range scores {
		if scores[i].TraitKey == TraitExtraversion {
			extraversion = &scores[i]
		}
	}
	if extraversion == nil {
		t.Fatal("se esperaba una puntuación para 'extraversion'")
	}
	if extraversion.AnsweredCount != 4 {
		t.Errorf("se esperaban 4 respuestas contabilizadas, hay %d", extraversion.AnsweredCount)
	}
	// (2+4+3+5)/4 = 3.5
	if extraversion.AverageScore < 3.49 || extraversion.AverageScore > 3.51 {
		t.Errorf("se esperaba una media ≈3.5, se obtuvo %v", extraversion.AverageScore)
	}
}

// --- Preferencias de pareja -----------------------------------------------

func TestPostgresRepository_PartnerPreferences(t *testing.T) {
	pool := testPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	profile := seedTestProfile(t, pool, repo)

	pp, err := repo.GetPartnerPreferences(ctx, profile.ID)
	if err != nil {
		t.Fatalf("GetPartnerPreferences (vacío): %v", err)
	}
	if pp.AgeMin != nil || pp.AgeMax != nil || pp.DesiredTraits != nil {
		t.Errorf("se esperaban todos los campos a nil antes de rellenar, se obtuvo: %+v", pp)
	}

	ageMin, ageMax := 38, 58
	pp, err = repo.UpsertPartnerPreferences(ctx, profile.ID, PartnerPreferencesPatch{
		AgeMin:        FieldOf(&ageMin),
		AgeMax:        FieldOf(&ageMax),
		DesiredTraits: FieldOf([]string{"humorous", "kind_hearted"}),
	})
	if err != nil {
		t.Fatalf("UpsertPartnerPreferences (primer patch): %v", err)
	}
	if pp.AgeMin == nil || *pp.AgeMin != 38 {
		t.Errorf("rango de edad no se guardó bien: %+v", pp)
	}

	fun := 5
	pp, err = repo.UpsertPartnerPreferences(ctx, profile.ID, PartnerPreferencesPatch{
		ImportanceFun: FieldOf(&fun),
	})
	if err != nil {
		t.Fatalf("UpsertPartnerPreferences (segundo patch): %v", err)
	}
	if pp.ImportanceFun == nil || *pp.ImportanceFun != 5 {
		t.Errorf("importance_fun no se guardó: %+v", pp)
	}
	if pp.AgeMin == nil || *pp.AgeMin != 38 {
		t.Errorf("el patch parcial no debería haber borrado age_min, se obtuvo: %+v", pp.AgeMin)
	}
}
