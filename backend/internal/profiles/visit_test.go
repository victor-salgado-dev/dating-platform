package profiles

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type fakeVisits struct{ calls [][2]uuid.UUID }

func (f *fakeVisits) Record(_ context.Context, visitor, visited uuid.UUID) error {
	f.calls = append(f.calls, [2]uuid.UUID{visitor, visited})
	return nil
}

func TestSkipVisit(t *testing.T) {
	cases := map[string]bool{
		"/api/v1/profiles/x/full":                    false,
		"/api/v1/profiles/x/full?visit=0":            true,
		"/api/v1/profiles/x/full?visit=1":            false,
		"/api/v1/profiles/x/full?visit=":             false,
		"/api/v1/profiles/x/full?other=0":            false,
		"/api/v1/profiles/x/full?size=thumb&visit=0": true,
	}
	for url, want := range cases {
		if got := skipVisit(httptest.NewRequest("GET", url, nil)); got != want {
			t.Errorf("skipVisit(%q) = %v, se esperaba %v", url, got, want)
		}
	}
}

// El registro de visitas depende solo de la opción explícita y de quién mira.
func TestGetFullPublicProfile_VisitRecording(t *testing.T) {
	viewerProfile, target := uuid.New(), uuid.New()

	cases := []struct {
		name      string
		repo      *fakeRepo
		profileID uuid.UUID
		opts      FullPublicOptions
		wantCalls int
	}{
		{"se registra por defecto", &fakeRepo{myProfileID: viewerProfile}, target, FullPublicOptions{}, 1},
		{"SkipVisit no registra (precarga)", &fakeRepo{myProfileID: viewerProfile}, target, FullPublicOptions{SkipVisit: true}, 0},
		{"mirar el propio perfil no cuenta", &fakeRepo{myProfileID: viewerProfile}, viewerProfile, FullPublicOptions{}, 0},
		{"visitante sin perfil no registra y no falla", &fakeRepo{noProfile: true}, target, FullPublicOptions{}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			visits := &fakeVisits{}
			svc := &Service{repo: tc.repo, visits: visits}

			full, err := svc.GetFullPublicProfile(context.Background(), uuid.New(), tc.profileID, tc.opts)
			if err != nil {
				t.Fatalf("GetFullPublicProfile: %v", err)
			}
			if full == nil || full.Profile == nil {
				t.Fatal("se esperaba el perfil completo")
			}
			if len(visits.calls) != tc.wantCalls {
				t.Fatalf("visitas registradas = %d, se esperaban %d", len(visits.calls), tc.wantCalls)
			}
			if tc.wantCalls == 1 && visits.calls[0] != [2]uuid.UUID{viewerProfile, target} {
				t.Errorf("la visita debe registrarse con IDs de PERFIL (visitante, visitado): %v", visits.calls[0])
			}
		})
	}
}
