//go:build integration

// Test de integración de la Fase 2: a diferencia de
// migration_000013_integration_test.go (que valida los CHECK/FK de la
// migración con SQL directo), este fichero ejercita el código Go de
// PostgresRepository tal cual lo usará el Service: construye el pool,
// llama a los métodos reales (UpsertProfileHobby, GetPartnerPreferences,
// etc.) y comprueba que lo que se guarda es lo que se lee.
//
// Cómo ejecutarlo (mismos requisitos que el test de la migración):
//
//	migrate -database "$TEST_DATABASE_URL" -path ./migrations up
//	TEST_DATABASE_URL="postgres://user:pass@localhost:5432/dating_platform_test?sslmode=disable" \
//	    go test -tags=integration ./internal/profiles/... -run TestPostgresRepository -v
package profiles

import (
	"context"
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
// propio PostgresRepository (no SQL a mano, para probar Create() de
// paso) y registra su borrado al final del test. Al borrar el usuario,
// el ON DELETE CASCADE se lleva perfil, hobbies, respuestas de
// personalidad y preferencias de pareja con él.
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

// --- Hobbies --------------------------------------------------------------

func TestPostgresRepository_ProfileHobbies(t *testing.T) {
	pool := testPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	profile := seedTestProfile(t, pool, repo)

	intensity := 4
	ph, err := repo.UpsertProfileHobby(ctx, profile.ID, "cooking_baking", true, &intensity)
	if err != nil {
		t.Fatalf("UpsertProfileHobby: %v", err)
	}
	if !ph.Liked || ph.Intensity == nil || *ph.Intensity != 4 {
		t.Errorf("resultado inesperado del upsert: %+v", ph)
	}

	list, err := repo.ListProfileHobbies(ctx, profile.ID)
	if err != nil {
		t.Fatalf("ListProfileHobbies: %v", err)
	}
	if len(list) != 1 || list[0].HobbyKey != "cooking_baking" {
		t.Fatalf("se esperaba 1 hobby 'cooking_baking', se obtuvo: %+v", list)
	}

	// Reescribir la misma respuesta (upsert): baja la intensidad a 2.
	lowerIntensity := 2
	if _, err := repo.UpsertProfileHobby(ctx, profile.ID, "cooking_baking", true, &lowerIntensity); err != nil {
		t.Fatalf("UpsertProfileHobby (update): %v", err)
	}
	list, err = repo.ListProfileHobbies(ctx, profile.ID)
	if err != nil {
		t.Fatalf("ListProfileHobbies tras actualizar: %v", err)
	}
	if len(list) != 1 || list[0].Intensity == nil || *list[0].Intensity != 2 {
		t.Fatalf("se esperaba intensidad actualizada a 2, se obtuvo: %+v", list)
	}

	// Cambiar de "me gusta" a "no me gusta": intensity debe quedar a nil.
	if _, err := repo.UpsertProfileHobby(ctx, profile.ID, "cooking_baking", false, nil); err != nil {
		t.Fatalf("UpsertProfileHobby (dislike): %v", err)
	}
	list, err = repo.ListProfileHobbies(ctx, profile.ID)
	if err != nil {
		t.Fatalf("ListProfileHobbies tras marcar dislike: %v", err)
	}
	if len(list) != 1 || list[0].Liked || list[0].Intensity != nil {
		t.Fatalf("se esperaba liked=false e intensity=nil, se obtuvo: %+v", list[0])
	}

	// Borrar: vuelve a "no contestado".
	if err := repo.DeleteProfileHobby(ctx, profile.ID, "cooking_baking"); err != nil {
		t.Fatalf("DeleteProfileHobby: %v", err)
	}
	list, err = repo.ListProfileHobbies(ctx, profile.ID)
	if err != nil {
		t.Fatalf("ListProfileHobbies tras borrar: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("se esperaba la lista vacía tras borrar, hay %d elementos", len(list))
	}

	// Borrar algo que no existe es idempotente, no debe fallar.
	if err := repo.DeleteProfileHobby(ctx, profile.ID, "cooking_baking"); err != nil {
		t.Errorf("borrar un hobby ya borrado no debería fallar: %v", err)
	}
}

func TestPostgresRepository_HobbyCatalog(t *testing.T) {
	pool := testPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	defs, err := repo.ListHobbyDefinitions(ctx)
	if err != nil {
		t.Fatalf("ListHobbyDefinitions: %v", err)
	}
	if len(defs) != 21 {
		t.Errorf("se esperaban 21 hobbies, se obtuvieron %d", len(defs))
	}
	// Debe venir ordenado por categoría: la primera categoría en orden
	// alfabético es "at_home".
	if len(defs) > 0 && defs[0].Category != "at_home" {
		t.Errorf("se esperaba que empezara por la categoría 'at_home', empezó por %q", defs[0].Category)
	}
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

	// Reescribir una respuesta (upsert) y comprobar que no duplica fila.
	if _, err := repo.UpsertPersonalityAnswer(ctx, profile.ID, "extra_reserved_calm", 1); err != nil {
		t.Fatalf("UpsertPersonalityAnswer (update): %v", err)
	}
	list, err = repo.ListPersonalityAnswers(ctx, profile.ID)
	if err != nil {
		t.Fatalf("ListPersonalityAnswers tras actualizar: %v", err)
	}
	if len(list) != 4 {
		t.Fatalf("se esperaban seguir siendo 4 respuestas tras el upsert, hay %d", len(list))
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
	// (1 + 4 + 3 + 5) / 4 = 3.25
	if extraversion.AverageScore < 3.24 || extraversion.AverageScore > 3.26 {
		t.Errorf("se esperaba una media ≈3.25, se obtuvo %v", extraversion.AverageScore)
	}
}

// --- Preferencias de pareja -----------------------------------------------

func TestPostgresRepository_PartnerPreferences(t *testing.T) {
	pool := testPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	profile := seedTestProfile(t, pool, repo)

	// Sin rellenar todavía: no es un error, todo viene a nil.
	pp, err := repo.GetPartnerPreferences(ctx, profile.ID)
	if err != nil {
		t.Fatalf("GetPartnerPreferences (vacío): %v", err)
	}
	if pp.AgeMin != nil || pp.AgeMax != nil || pp.DesiredTraits != nil {
		t.Errorf("se esperaban todos los campos a nil antes de rellenar, se obtuvo: %+v", pp)
	}

	ageMin, ageMax := 38, 58
	pp, err = repo.UpsertPartnerPreferences(ctx, profile.ID, PartnerPreferencesPatch{
		AgeMinSet: true, AgeMin: &ageMin,
		AgeMaxSet: true, AgeMax: &ageMax,
		DesiredTraitsSet: true, DesiredTraits: []string{"humorous", "kind_hearted"},
	})
	if err != nil {
		t.Fatalf("UpsertPartnerPreferences (primer patch): %v", err)
	}
	if pp.AgeMin == nil || *pp.AgeMin != 38 || pp.AgeMax == nil || *pp.AgeMax != 58 {
		t.Errorf("rango de edad no se guardó bien: %+v", pp)
	}
	if len(pp.DesiredTraits) != 2 {
		t.Errorf("desired_traits no se guardó bien: %+v", pp.DesiredTraits)
	}

	// Segundo patch parcial: solo toca una importancia. Los campos ya
	// guardados (age_min, age_max, desired_traits) deben seguir intactos
	// — el upsert no debe pisarlos con NULL.
	fun := 5
	pp, err = repo.UpsertPartnerPreferences(ctx, profile.ID, PartnerPreferencesPatch{
		ImportanceFunSet: true, ImportanceFun: &fun,
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
	if len(pp.DesiredTraits) != 2 {
		t.Errorf("el patch parcial no debería haber borrado desired_traits, se obtuvo: %+v", pp.DesiredTraits)
	}

	// Borrar explícitamente un campo (patch a nil con el flag activado).
	pp, err = repo.UpsertPartnerPreferences(ctx, profile.ID, PartnerPreferencesPatch{
		AgeMinSet: true, AgeMin: nil,
	})
	if err != nil {
		t.Fatalf("UpsertPartnerPreferences (borrar age_min): %v", err)
	}
	if pp.AgeMin != nil {
		t.Errorf("age_min debería haber quedado a nil, se obtuvo: %v", *pp.AgeMin)
	}
	if pp.AgeMax == nil || *pp.AgeMax != 58 {
		t.Errorf("age_max no debería haberse visto afectado: %+v", pp.AgeMax)
	}
}
