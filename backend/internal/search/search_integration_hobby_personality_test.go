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

// NOTA: aquí vivían los tests de filtrado por hobbies (Fase 2). Se
// retiraron porque la escritura de hobbies ya no existe en profiles
// (fue sustituida por profile_interests): sin un camino de escritura, los
// tests no pueden pasar. El filtro de hobbies que aún queda en el
// repositorio de search (hobbyExistsClause) consulta la tabla
// profile_hobbies, que ya no se alimenta; es candidato a limpieza aparte.
// Se conservan los tests de personalidad, que sí siguen vigentes.

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
