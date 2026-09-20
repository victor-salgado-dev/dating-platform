//go:build integration

package search_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
	"dating-platform/backend/internal/search"
	"dating-platform/backend/internal/testutil"
	"dating-platform/backend/internal/users"
)

// Cambiado wantsChildren de *bool a *string
func createUserWithProfile(t *testing.T, ctx context.Context, usersRepo users.Repository, profilesRepo profiles.Repository, wantsChildren *string) uuid.UUID {
	t.Helper()

	u := &users.User{Email: testutil.UniqueEmail(), PasswordHash: "x"}
	if err := usersRepo.Create(ctx, u); err != nil {
		t.Fatalf("crear usuario de prueba: %v", err)
	}

	p := &profiles.Profile{
		UserID:        u.ID,
		DisplayName:   "Persona de prueba",
		BirthDate:     time.Now().AddDate(-30, 0, 0),
		Gender:        profiles.GenderOther,
		CountryCode:   "ES",
		WantsChildren: wantsChildren,
	}
	if err := profilesRepo.Create(ctx, p); err != nil {
		t.Fatalf("crear perfil de prueba: %v", err)
	}
	return p.ID
}

func containsProfile(items []search.ResultItem, id uuid.UUID) bool {
	for _, it := range items {
		if it.ProfileID == id {
			return true
		}
	}
	return false
}

// TestIntegration_Search_MissingDataRule verifica contra PostgreSQL real
// la regla fundamental de la sección 8 con el nuevo formato string ("yes").
func TestIntegration_Search_MissingDataRule(t *testing.T) {
	pool := testutil.RequireDB(t)
	ctx := context.Background()

	usersRepo := users.NewPostgresRepository(pool)
	profilesRepo := profiles.NewPostgresRepository(pool)
	searchRepo := search.NewPostgresRepository(pool)

	searcher := &users.User{Email: testutil.UniqueEmail(), PasswordHash: "x"}
	if err := usersRepo.Create(ctx, searcher); err != nil {
		t.Fatalf("crear usuario buscador: %v", err)
	}

	yes := "yes"
	withData := createUserWithProfile(t, ctx, usersRepo, profilesRepo, &yes)
	withoutData := createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)

	// Sin filtro: ambos perfiles deben aparecer
	resultAll, err := searchRepo.Search(ctx, search.Params{
		ExcludeUserID: searcher.ID, Sort: search.SortRecent, Page: 1, PageSize: 50,
	})
	if err != nil {
		t.Fatalf("Search sin filtro: %v", err)
	}
	if !containsProfile(resultAll.Items, withData) {
		t.Error("sin filtro, el perfil que SÍ indicó el dato debería aparecer")
	}
	if !containsProfile(resultAll.Items, withoutData) {
		t.Error("sin filtro, el perfil que NO indicó el dato también debería aparecer")
	}

	// Con filtro wants_children="yes": SOLO quien lo indicó explícitamente
	yesVal := "yes"
	resultFiltered, err := searchRepo.Search(ctx, search.Params{
		ExcludeUserID: searcher.ID, Sort: search.SortRecent, Page: 1, PageSize: 50,
		Filters: search.Filters{WantsChildren: &yesVal},
	})
	if err != nil {
		t.Fatalf("Search con filtro wants_children=yes: %v", err)
	}
	if !containsProfile(resultFiltered.Items, withData) {
		t.Error("el perfil que indicó wants_children=yes debería aparecer al filtrar por ese valor")
	}
	if containsProfile(resultFiltered.Items, withoutData) {
		t.Error("REGLA DE DATOS FALTANTES: un perfil que no indicó wants_children NO debería aparecer al filtrar por ese campo")
	}
}

func TestIntegration_Search_ExcludesSelf(t *testing.T) {
	pool := testutil.RequireDB(t)
	ctx := context.Background()

	usersRepo := users.NewPostgresRepository(pool)
	profilesRepo := profiles.NewPostgresRepository(pool)
	searchRepo := search.NewPostgresRepository(pool)

	u := &users.User{Email: testutil.UniqueEmail(), PasswordHash: "x"}
	if err := usersRepo.Create(ctx, u); err != nil {
		t.Fatalf("crear usuario: %v", err)
	}
	createUserWithProfile(t, ctx, usersRepo, profilesRepo, nil)

	p := &profiles.Profile{
		UserID:      u.ID,
		DisplayName: "Yo mismo",
		BirthDate:   time.Now().AddDate(-25, 0, 0),
		Gender:      profiles.GenderOther,
		CountryCode: "ES",
	}
	if err := profilesRepo.Create(ctx, p); err != nil {
		t.Fatalf("crear el propio perfil: %v", err)
	}

	result, err := searchRepo.Search(ctx, search.Params{
		ExcludeUserID: u.ID, Sort: search.SortRecent, Page: 1, PageSize: 50,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if containsProfile(result.Items, p.ID) {
		t.Error("una búsqueda nunca debería devolver el propio perfil de quien busca")
	}
}