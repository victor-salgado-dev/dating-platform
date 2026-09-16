//go:build integration

package likes

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"

    "dating-platform/backend/internal/profiles"
    "dating-platform/backend/internal/testutil"
    "dating-platform/backend/internal/users"
)

func createIntegrationProfile(t *testing.T, ctx context.Context, usersRepo users.Repository, profilesRepo profiles.Repository) uuid.UUID {
    t.Helper()
    user := &users.User{Email: testutil.UniqueEmail(), PasswordHash: "x"}
    if err := usersRepo.Create(ctx, user); err != nil { t.Fatalf("crear usuario: %v", err) }
    profile := &profiles.Profile{UserID: user.ID, DisplayName: "Persona like", BirthDate: time.Now().AddDate(-30, 0, 0), Gender: profiles.GenderOther, CountryCode: "ES"}
    if err := profilesRepo.Create(ctx, profile); err != nil { t.Fatalf("crear perfil: %v", err) }
    return profile.ID
}

func TestIntegrationMutualLikeCreatesAndRemovesMatch(t *testing.T) {
    pool := testutil.RequireDB(t)
    ctx := context.Background()
    usersRepo := users.NewPostgresRepository(pool)
    profilesRepo := profiles.NewPostgresRepository(pool)
    repo := NewPostgresRepository(pool)
    profileA := createIntegrationProfile(t, ctx, usersRepo, profilesRepo)
    profileB := createIntegrationProfile(t, ctx, usersRepo, profilesRepo)

    matched, err := repo.Add(ctx, profileA, profileB)
    if err != nil || matched { t.Fatalf("primer like = (%v, %v), want (false, nil)", matched, err) }
    matched, err = repo.Add(ctx, profileB, profileA)
    if err != nil || !matched { t.Fatalf("like mutuo = (%v, %v), want (true, nil)", matched, err) }

    result, err := repo.ListMatches(ctx, profileA, 1, 1)
    if err != nil || result.Total != 1 || len(result.Items) != 1 { t.Fatalf("listar matches = (%v, %v), want one result", result, err) }
    if ok, err := repo.HasMatch(ctx, profileA, profileB); err != nil || !ok { t.Fatalf("HasMatch = (%v, %v), want true", ok, err) }

    if err := repo.Remove(ctx, profileA, profileB); err != nil { t.Fatalf("quitar like: %v", err) }
    if ok, err := repo.HasMatch(ctx, profileA, profileB); err != nil || ok { t.Fatalf("HasMatch tras unlike = (%v, %v), want false", ok, err) }
}
