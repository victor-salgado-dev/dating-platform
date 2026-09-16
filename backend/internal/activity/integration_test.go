//go:build integration

package activity

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/blocking"
	"dating-platform/backend/internal/favorites"
	"dating-platform/backend/internal/likes"
	"dating-platform/backend/internal/profiles"
	"dating-platform/backend/internal/testutil"
	"dating-platform/backend/internal/users"
)

type integrationProfile struct{ userID, profileID uuid.UUID }

func newIntegrationProfile(t *testing.T, ctx context.Context, usersRepo users.Repository, profilesRepo profiles.Repository) integrationProfile {
	t.Helper()
	user := &users.User{Email: testutil.UniqueEmail(), PasswordHash: "x"}
	if err := usersRepo.Create(ctx, user); err != nil {
		t.Fatalf("crear usuario: %v", err)
	}
	profile := &profiles.Profile{UserID: user.ID, DisplayName: "Actividad", BirthDate: time.Now().AddDate(-30, 0, 0), Gender: profiles.GenderOther, CountryCode: "ES"}
	if err := profilesRepo.Create(ctx, profile); err != nil {
		t.Fatalf("crear perfil: %v", err)
	}
	return integrationProfile{userID: user.ID, profileID: profile.ID}
}

func TestIntegrationActivityIncludesLikeMatchAndFavorite(t *testing.T) {
	pool := testutil.RequireDB(t)
	ctx := context.Background()
	usersRepo := users.NewPostgresRepository(pool)
	profilesRepo := profiles.NewPostgresRepository(pool)
	favoritesRepo := favorites.NewPostgresRepository(pool)
	likesRepo := likes.NewPostgresRepository(pool)
	activityRepo := NewPostgresRepository(pool)

	viewer := newIntegrationProfile(t, ctx, usersRepo, profilesRepo)
	liker := newIntegrationProfile(t, ctx, usersRepo, profilesRepo)
	matcher := newIntegrationProfile(t, ctx, usersRepo, profilesRepo)
	favoritist := newIntegrationProfile(t, ctx, usersRepo, profilesRepo)
	blocked := newIntegrationProfile(t, ctx, usersRepo, profilesRepo)

	if _, err := likesRepo.Add(ctx, liker.profileID, viewer.profileID); err != nil {
		t.Fatalf("like recibido: %v", err)
	}
	if _, err := likesRepo.Add(ctx, viewer.profileID, matcher.profileID); err != nil {
		t.Fatalf("like del viewer: %v", err)
	}
	if _, err := likesRepo.Add(ctx, matcher.profileID, viewer.profileID); err != nil {
		t.Fatalf("match: %v", err)
	}
	if err := favoritesRepo.Add(ctx, favoritist.userID, viewer.profileID); err != nil {
		t.Fatalf("favorito recibido: %v", err)
	}
	if err := blocking.NewPostgresRepository(pool).Add(ctx, viewer.userID, blocked.userID); err != nil {
		t.Fatalf("bloquear perfil: %v", err)
	}
	if _, err := likesRepo.Add(ctx, blocked.profileID, viewer.profileID); err != nil {
		t.Fatalf("like bloqueado: %v", err)
	}

	result, err := activityRepo.List(ctx, viewer.profileID, 1, 10)
	if err != nil {
		t.Fatalf("listar actividad: %v", err)
	}
	if result.Total < 3 {
		t.Fatalf("total de actividad = %d, esperados al menos 3", result.Total)
	}

	seen := map[string]bool{}
	for i, item := range result.Items {
		seen[item.EventType] = true
		if i > 0 && result.Items[i-1].CreatedAt.Before(item.CreatedAt) {
			t.Fatalf("actividad no ordenada por fecha descendente")
		}
	}
	for _, eventType := range []string{EventLikeReceived, EventMatchCreated, EventFavoriteReceived} {
		if !seen[eventType] {
			t.Errorf("falta evento %s", eventType)
		}
	}
	for _, item := range result.Items {
		if item.ProfileID == blocked.profileID {
			t.Fatalf("un evento bloqueado no debería aparecer")
		}
	}
}
