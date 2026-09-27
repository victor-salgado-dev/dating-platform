//go:build integration

// Reutiliza createUserWithProfile y containsProfile, definidos en
// integration_test.go dentro de este mismo paquete (search_test).
package search_test

import (
	"context"
	"testing"

	"dating-platform/backend/internal/profiles"
	"dating-platform/backend/internal/search"
	"dating-platform/backend/internal/testutil"
	"dating-platform/backend/internal/users"
)

// Los tests de filtrado por hobbies que vivían aquí se adaptaron a la
// API nueva de intereses (Fase 2 -> profile_interests): pertenencia
// simple (TestIntegration_Search_InterestFilter) y, para intereses con
// has_level=true, pertenencia + rango de nivel
// (TestIntegration_Search_InterestLevelFilter_MissingDataRule) — mismo
// comparador </o> que antes tenían los hobbies.

func TestIntegration_Search_InterestFilter(t *testing.T) {
	pool := testutil.RequireDB(t)
	ctx := context.Background()

	usersRepo := users.NewPostgresRepository(pool)
	profilesRepo := profiles.NewPostgresRepository(pool)
	searchRepo := search.NewPostgresRepository(pool)

	// Cogemos dos claves reales del catálogo en vez de hardcodearlas: así
	// no dependemos de qué intereses siembra la migración, y la FK de
	// profile_interests nunca nos rechazará una clave inventada.
	defs, err := profilesRepo.ListInterestDefinitions(ctx)
	if err != nil {
		t.Fatalf("listar catálogo de intereses: %v", err)
	}
	if len(defs) < 2 {
		t.Fatalf("el catálogo de intereses necesita al menos 2 entradas, tiene %d", len(defs))
	}
	interestA := defs[0].Key
	interestB := defs[1].Key

	searcher := &users.User{Email: testutil.UniqueEmail(), PasswordHash: "x"}
	if err := usersRepo.Create(ctx, searcher); err != nil {
		t.Fatalf("crear usuario buscador: %v", err)
	}

	// withLevel tiene el interés A con un nivel concreto; withoutLevel lo
	// tiene con level a NULL. El filtro actual solo mira pertenencia, así
	// que ambos deben aparecer.
	withLevel := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	level5 := 5
	if _, err := profilesRepo.UpsertProfileInterest(ctx, withLevel, interestA, &level5); err != nil {
		t.Fatalf("añadir interés A (con nivel): %v", err)
	}

	withoutLevel := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertProfileInterest(ctx, withoutLevel, interestA, nil); err != nil {
		t.Fatalf("añadir interés A (sin nivel): %v", err)
	}

	// otherInterest solo tiene el interés B: no debe aparecer al filtrar por A.
	otherInterest := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertProfileInterest(ctx, otherInterest, interestB, nil); err != nil {
		t.Fatalf("añadir interés B: %v", err)
	}

	// noInterest no declara ninguno.
	noInterest := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)

	result, err := searchRepo.Search(ctx, search.Params{
		ExcludeUserID: searcher.ID, Sort: search.SortRecent, Page: 1, PageSize: 50,
		Filters: search.Filters{Interests: []search.InterestFilter{{Key: interestA}}},
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if !containsProfile(result.Items, withLevel) {
		t.Error("un perfil con el interés A (con nivel) debería aparecer al filtrar por A")
	}
	if !containsProfile(result.Items, withoutLevel) {
		t.Error("un perfil con el interés A (level NULL) debería aparecer al filtrar por A")
	}
	if containsProfile(result.Items, otherInterest) {
		t.Error("un perfil con solo el interés B NO debería aparecer al filtrar por A")
	}
	if containsProfile(result.Items, noInterest) {
		t.Error("REGLA DE DATOS FALTANTES: un perfil sin ese interés no debería aparecer al filtrar")
	}
}

// TestIntegration_Search_InterestLevelFilter_MissingDataRule prueba el
// filtro de nivel (?interest_x_min=/_max=) que usan los intereses con
// has_level=true: además de tener el interés marcado, level tiene que
// caer en el rango pedido. Un perfil con el interés pero sin level
// (NULL) nunca cumple un filtro con rango — regla de datos faltantes.
func TestIntegration_Search_InterestLevelFilter_MissingDataRule(t *testing.T) {
	pool := testutil.RequireDB(t)
	ctx := context.Background()

	usersRepo := users.NewPostgresRepository(pool)
	profilesRepo := profiles.NewPostgresRepository(pool)
	searchRepo := search.NewPostgresRepository(pool)

	defs, err := profilesRepo.ListInterestDefinitions(ctx)
	if err != nil {
		t.Fatalf("listar catálogo de intereses: %v", err)
	}
	var leveled *profiles.InterestDefinition
	for i := range defs {
		if defs[i].HasLevel {
			leveled = &defs[i]
			break
		}
	}
	if leveled == nil {
		t.Skip("el catálogo de intereses no tiene ninguna entrada con has_level=true todavía")
	}

	searcher := &users.User{Email: testutil.UniqueEmail(), PasswordHash: "x"}
	if err := usersRepo.Create(ctx, searcher); err != nil {
		t.Fatalf("crear usuario buscador: %v", err)
	}

	highLevel := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	five := 5
	if _, err := profilesRepo.UpsertProfileInterest(ctx, highLevel, leveled.Key, &five); err != nil {
		t.Fatalf("añadir interés con nivel 5: %v", err)
	}

	lowLevel := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	one := 1
	if _, err := profilesRepo.UpsertProfileInterest(ctx, lowLevel, leveled.Key, &one); err != nil {
		t.Fatalf("añadir interés con nivel 1: %v", err)
	}

	// noLevel tiene el interés marcado pero sin nivel (NULL).
	noLevel := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertProfileInterest(ctx, noLevel, leveled.Key, nil); err != nil {
		t.Fatalf("añadir interés sin nivel: %v", err)
	}

	four := 4
	result, err := searchRepo.Search(ctx, search.Params{
		ExcludeUserID: searcher.ID, Sort: search.SortRecent, Page: 1, PageSize: 50,
		Filters: search.Filters{Interests: []search.InterestFilter{{Key: leveled.Key, Min: &four}}},
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if !containsProfile(result.Items, highLevel) {
		t.Error("un perfil con nivel 5 debería aparecer al filtrar por nivel >=4")
	}
	if containsProfile(result.Items, lowLevel) {
		t.Error("un perfil con nivel 1 NO debería aparecer al filtrar por nivel >=4")
	}
	if containsProfile(result.Items, noLevel) {
		t.Error("REGLA DE DATOS FALTANTES: un perfil con el interés marcado pero sin nivel no debería aparecer al filtrar por nivel")
	}
}

func TestIntegration_Search_PersonalityFilter_MissingDataRule(t *testing.T) {
	pool := testutil.RequireDB(t)
	ctx := context.Background()

	usersRepo := users.NewPostgresRepository(pool)
	profilesRepo := profiles.NewPostgresRepository(pool)
	searchRepo := search.NewPostgresRepository(pool)

	searcher := &users.User{Email: testutil.UniqueEmail(), PasswordHash: "x"}
	if err := usersRepo.Create(ctx, searcher); err != nil {
		t.Fatalf("crear usuario buscador: %v", err)
	}

	highScore := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertPersonalityAnswer(ctx, highScore, "extra_funny_laughs", 5); err != nil {
		t.Fatalf("responder personalidad (highScore): %v", err)
	}

	lowScore := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertPersonalityAnswer(ctx, lowScore, "extra_reserved_calm", 1); err != nil {
		t.Fatalf("responder personalidad (lowScore): %v", err)
	}

	// noAnswer no contesta ninguna afirmación de personalidad.
	noAnswer := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)

	min := 4.0
	result, err := searchRepo.Search(ctx, search.Params{
		ExcludeUserID: searcher.ID, Sort: search.SortRecent, Page: 1, PageSize: 50,
		Filters: search.Filters{PersonalityTraits: []search.PersonalityFilter{
			{TraitKey: "extraversion", Min: &min},
		}},
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if !containsProfile(result.Items, highScore) {
		t.Error("un perfil con media 5 en extraversión debería aparecer al filtrar >=4")
	}
	if containsProfile(result.Items, lowScore) {
		t.Error("un perfil con media 1 en extraversión NO debería aparecer al filtrar >=4")
	}
	if containsProfile(result.Items, noAnswer) {
		t.Error("REGLA DE DATOS FALTANTES: un perfil sin ninguna respuesta de personalidad no debería aparecer al filtrar por un rasgo")
	}
}
