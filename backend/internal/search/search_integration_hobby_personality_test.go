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

func intensity(n int) *int { return &n }

// --- Hobbies --------------------------------------------------------------

func TestIntegration_Search_HobbyFilter_MissingDataRule(t *testing.T) {
	pool := testutil.RequireDB(t)
	ctx := context.Background()

	usersRepo := users.NewPostgresRepository(pool)
	profilesRepo := profiles.NewPostgresRepository(pool)
	searchRepo := search.NewPostgresRepository(pool)

	searcher := &users.User{Email: testutil.UniqueEmail(), PasswordHash: "x"}
	if err := usersRepo.Create(ctx, searcher); err != nil {
		t.Fatalf("crear usuario buscador: %v", err)
	}

	likesALot := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertProfileHobby(ctx, likesALot, "traveling", true, intensity(4)); err != nil {
		t.Fatalf("marcar hobby (likesALot): %v", err)
	}

	likesALittle := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertProfileHobby(ctx, likesALittle, "traveling", true, intensity(2)); err != nil {
		t.Fatalf("marcar hobby (likesALittle): %v", err)
	}

	dislikes := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertProfileHobby(ctx, dislikes, "traveling", false, nil); err != nil {
		t.Fatalf("marcar hobby (dislikes): %v", err)
	}

	// noAnswer no contesta nada sobre 'traveling'.
	noAnswer := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)

	// Sin filtro: los 4 perfiles deberían aparecer, filtren o no filtren
	// por 'traveling' — la regla de datos faltantes solo entra en juego
	// cuando el filtro se aplica.
	all, err := searchRepo.Search(ctx, search.Params{
		ExcludeUserID: searcher.ID, Sort: search.SortRecent, Page: 1, PageSize: 50,
	})
	if err != nil {
		t.Fatalf("Search sin filtro: %v", err)
	}
	if !containsProfile(all.Items, likesALot) {
		t.Error("sin filtro, likesALot debería aparecer")
	}
	if !containsProfile(all.Items, likesALittle) {
		t.Error("sin filtro, likesALittle debería aparecer")
	}
	if !containsProfile(all.Items, dislikes) {
		t.Error("sin filtro, dislikes debería aparecer")
	}
	if !containsProfile(all.Items, noAnswer) {
		t.Error("sin filtro, noAnswer debería aparecer")
	}

	// Filtro hobby_traveling_min=3: solo likesALot (intensidad 4) cumple.
	filtered, err := searchRepo.Search(ctx, search.Params{
		ExcludeUserID: searcher.ID, Sort: search.SortRecent, Page: 1, PageSize: 50,
		Filters: search.Filters{Hobbies: []search.HobbyFilter{{Key: "traveling", Min: intensity(3)}}},
	})
	if err != nil {
		t.Fatalf("Search con filtro de hobby: %v", err)
	}

	if !containsProfile(filtered.Items, likesALot) {
		t.Error("el perfil con intensidad 4 debería aparecer al filtrar >=3")
	}
	if containsProfile(filtered.Items, likesALittle) {
		t.Error("el perfil con intensidad 2 NO debería aparecer al filtrar >=3")
	}
	if containsProfile(filtered.Items, dislikes) {
		t.Error("REGLA DE DATOS FALTANTES: un perfil que dijo 'no me gusta' no debería aparecer al filtrar por intensidad")
	}
	if containsProfile(filtered.Items, noAnswer) {
		t.Error("REGLA DE DATOS FALTANTES: un perfil que no contestó ese hobby no debería aparecer al filtrar por él")
	}
}

func TestIntegration_Search_HobbyFilter_BareKeyMeansLikedAnyIntensity(t *testing.T) {
	pool := testutil.RequireDB(t)
	ctx := context.Background()

	usersRepo := users.NewPostgresRepository(pool)
	profilesRepo := profiles.NewPostgresRepository(pool)
	searchRepo := search.NewPostgresRepository(pool)

	searcher := &users.User{Email: testutil.UniqueEmail(), PasswordHash: "x"}
	if err := usersRepo.Create(ctx, searcher); err != nil {
		t.Fatalf("crear usuario buscador: %v", err)
	}

	likesALittle := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertProfileHobby(ctx, likesALittle, "gardening", true, intensity(1)); err != nil {
		t.Fatalf("marcar hobby: %v", err)
	}
	dislikes := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertProfileHobby(ctx, dislikes, "gardening", false, nil); err != nil {
		t.Fatalf("marcar hobby: %v", err)
	}

	// Filtro sin Min/Max: "le gusta la jardinería", cualquier intensidad.
	result, err := searchRepo.Search(ctx, search.Params{
		ExcludeUserID: searcher.ID, Sort: search.SortRecent, Page: 1, PageSize: 50,
		Filters: search.Filters{Hobbies: []search.HobbyFilter{{Key: "gardening"}}},
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if !containsProfile(result.Items, likesALittle) {
		t.Error("sin min/max, cualquier intensidad de 'me gusta' debería aparecer")
	}
	if containsProfile(result.Items, dislikes) {
		t.Error("'no me gusta' no debería aparecer aunque el filtro no pida una intensidad mínima")
	}
}

func TestIntegration_Search_MultipleHobbyFiltersCombineWithAND(t *testing.T) {
	pool := testutil.RequireDB(t)
	ctx := context.Background()

	usersRepo := users.NewPostgresRepository(pool)
	profilesRepo := profiles.NewPostgresRepository(pool)
	searchRepo := search.NewPostgresRepository(pool)

	searcher := &users.User{Email: testutil.UniqueEmail(), PasswordHash: "x"}
	if err := usersRepo.Create(ctx, searcher); err != nil {
		t.Fatalf("crear usuario buscador: %v", err)
	}

	both := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertProfileHobby(ctx, both, "traveling", true, intensity(4)); err != nil {
		t.Fatalf("marcar traveling (both): %v", err)
	}
	if _, err := profilesRepo.UpsertProfileHobby(ctx, both, "gardening", true, intensity(2)); err != nil {
		t.Fatalf("marcar gardening (both): %v", err)
	}

	// onlyOne cumple 'traveling' pero no contesta 'gardening' en absoluto.
	onlyOne := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)
	if _, err := profilesRepo.UpsertProfileHobby(ctx, onlyOne, "traveling", true, intensity(4)); err != nil {
		t.Fatalf("marcar traveling (onlyOne): %v", err)
	}

	// jardinería<=3 (viajes>=3) equivale al ejemplo original: "jardinería<4 y viajes>3".
	result, err := searchRepo.Search(ctx, search.Params{
		ExcludeUserID: searcher.ID, Sort: search.SortRecent, Page: 1, PageSize: 50,
		Filters: search.Filters{Hobbies: []search.HobbyFilter{
			{Key: "traveling", Min: intensity(3)},
			{Key: "gardening", Max: intensity(3)},
		}},
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if !containsProfile(result.Items, both) {
		t.Error("el perfil que cumple ambos filtros debería aparecer")
	}
	if containsProfile(result.Items, onlyOne) {
		t.Error("los filtros de hobby se combinan con AND: cumplir solo uno no debería bastar")
	}
}

// --- Personalidad -----------------------------------------------------

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
