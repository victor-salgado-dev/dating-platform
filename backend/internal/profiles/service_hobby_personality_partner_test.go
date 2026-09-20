package profiles

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
)

// =====================================================================
// Fakes: un Repository y un Storage en memoria para poder probar la
// lógica de negocio de Service sin tocar una base de datos real. Solo
// implementan lo mínimo que necesitan estos tests; el resto de métodos
// existen únicamente para satisfacer la interfaz.
// =====================================================================

type fakeRepository struct {
	profile *Profile
	getErr  error

	upsertHobbyCalls []ProfileHobby
	upsertHobbyErr   error

	upsertPersonalityCalls []ProfilePersonalityAnswer
	upsertPersonalityErr   error

	upsertPartnerCalls []PartnerPreferencesPatch
	partnerPrefs       *PartnerPreferences
	upsertPartnerErr   error
}

func (f *fakeRepository) Create(ctx context.Context, p *Profile) error { return nil }

func (f *fakeRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.profile, nil
}

func (f *fakeRepository) GetPublicByID(ctx context.Context, id, viewerUserID uuid.UUID) (*Profile, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.profile, nil
}

func (f *fakeRepository) GetByIDAny(ctx context.Context, id uuid.UUID) (*Profile, error) {
	return f.profile, nil
}

func (f *fakeRepository) Update(ctx context.Context, userID uuid.UUID, patch ProfilePatch) (*Profile, error) {
	return f.profile, nil
}

func (f *fakeRepository) AddPhoto(ctx context.Context, profileID uuid.UUID, photo *Photo) error {
	return nil
}
func (f *fakeRepository) ListPhotos(ctx context.Context, profileID uuid.UUID) ([]Photo, error) {
	return nil, nil
}
func (f *fakeRepository) CountPhotos(ctx context.Context, profileID uuid.UUID) (int, error) {
	return 0, nil
}
func (f *fakeRepository) GetPhoto(ctx context.Context, profileID, photoID uuid.UUID) (*Photo, error) {
	return nil, ErrPhotoNotFound
}
func (f *fakeRepository) DeletePhoto(ctx context.Context, profileID, photoID uuid.UUID) error {
	return nil
}

func (f *fakeRepository) ListHobbyDefinitions(ctx context.Context) ([]HobbyDefinition, error) {
	return nil, nil
}
func (f *fakeRepository) ListPersonalityStatements(ctx context.Context) ([]PersonalityStatement, error) {
	return nil, nil
}

func (f *fakeRepository) UpsertProfileHobby(ctx context.Context, profileID uuid.UUID, hobbyKey string, liked bool, intensity *int) (*ProfileHobby, error) {
	if f.upsertHobbyErr != nil {
		return nil, f.upsertHobbyErr
	}
	ph := ProfileHobby{ProfileID: profileID, HobbyKey: hobbyKey, Liked: liked, Intensity: intensity, UpdatedAt: time.Now()}
	f.upsertHobbyCalls = append(f.upsertHobbyCalls, ph)
	return &ph, nil
}
func (f *fakeRepository) ListProfileHobbies(ctx context.Context, profileID uuid.UUID) ([]ProfileHobby, error) {
	return nil, nil
}
func (f *fakeRepository) DeleteProfileHobby(ctx context.Context, profileID uuid.UUID, hobbyKey string) error {
	return nil
}

func (f *fakeRepository) UpsertPersonalityAnswer(ctx context.Context, profileID uuid.UUID, statementKey string, score int) (*ProfilePersonalityAnswer, error) {
	if f.upsertPersonalityErr != nil {
		return nil, f.upsertPersonalityErr
	}
	a := ProfilePersonalityAnswer{ProfileID: profileID, StatementKey: statementKey, Score: score, UpdatedAt: time.Now()}
	f.upsertPersonalityCalls = append(f.upsertPersonalityCalls, a)
	return &a, nil
}
func (f *fakeRepository) ListPersonalityAnswers(ctx context.Context, profileID uuid.UUID) ([]ProfilePersonalityAnswer, error) {
	return nil, nil
}
func (f *fakeRepository) GetPersonalityTraitScores(ctx context.Context, profileID uuid.UUID) ([]PersonalityTraitScore, error) {
	return nil, nil
}

func (f *fakeRepository) GetPartnerPreferences(ctx context.Context, profileID uuid.UUID) (*PartnerPreferences, error) {
	if f.partnerPrefs != nil {
		return f.partnerPrefs, nil
	}
	return &PartnerPreferences{ProfileID: profileID}, nil
}
func (f *fakeRepository) UpsertPartnerPreferences(ctx context.Context, profileID uuid.UUID, patch PartnerPreferencesPatch) (*PartnerPreferences, error) {
	if f.upsertPartnerErr != nil {
		return nil, f.upsertPartnerErr
	}
	f.upsertPartnerCalls = append(f.upsertPartnerCalls, patch)
	return &PartnerPreferences{ProfileID: profileID}, nil
}

var _ Repository = (*fakeRepository)(nil)

// fakeStorage no se usa en estos tests (no tocan fotos), pero Service
// lo exige en el constructor.
type fakeStorage struct{}

func (fakeStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) { return nil, nil }
func (fakeStorage) Save(ctx context.Context, key string, r io.Reader) error     { return nil }
func (fakeStorage) Delete(ctx context.Context, key string) error                { return nil }

func newTestService(repo *fakeRepository) *Service {
	if repo.profile == nil {
		repo.profile = &Profile{ID: uuid.New(), UserID: uuid.New()}
	}
	return NewService(repo, fakeStorage{})
}

func intPtr(i int) *int { return &i }

// =====================================================================
// SetHobby
// =====================================================================

func TestSetHobby(t *testing.T) {
	t.Run("me gusta con intensidad válida se acepta y llega al repositorio", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		ph, err := svc.SetHobby(context.Background(), uuid.New(), "cooking_baking", true, intPtr(4))
		if err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
		if ph.HobbyKey != "cooking_baking" || !ph.Liked || ph.Intensity == nil || *ph.Intensity != 4 {
			t.Errorf("resultado inesperado: %+v", ph)
		}
		if len(repo.upsertHobbyCalls) != 1 {
			t.Errorf("se esperaba 1 llamada al repositorio, hubo %d", len(repo.upsertHobbyCalls))
		}
	})

	t.Run("no me gusta sin intensidad se acepta", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		_, err := svc.SetHobby(context.Background(), uuid.New(), "motorsport", false, nil)
		if err != nil {
			t.Errorf("no se esperaba error: %v", err)
		}
	})

	t.Run("no me gusta pero con intensidad se rechaza antes de llegar al repositorio", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		_, err := svc.SetHobby(context.Background(), uuid.New(), "gardening", false, intPtr(3))
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("se esperaba un ValidationError, se obtuvo: %v", err)
		}
		if valErr.Field != "intensity" {
			t.Errorf("se esperaba el campo 'intensity', fue %q", valErr.Field)
		}
		if len(repo.upsertHobbyCalls) != 0 {
			t.Error("no debería haber llegado a llamar al repositorio")
		}
	})

	for _, bad := range []int{0, 6, -1} {
		t.Run("intensidad fuera de rango se rechaza", func(t *testing.T) {
			repo := &fakeRepository{}
			svc := newTestService(repo)

			_, err := svc.SetHobby(context.Background(), uuid.New(), "reading", true, intPtr(bad))
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("intensidad %d: se esperaba ValidationError, se obtuvo: %v", bad, err)
			}
		})
	}

	t.Run("perfil inexistente propaga ErrNotFound", func(t *testing.T) {
		repo := &fakeRepository{getErr: ErrNotFound}
		svc := newTestService(repo)

		_, err := svc.SetHobby(context.Background(), uuid.New(), "reading", true, intPtr(3))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("se esperaba ErrNotFound, se obtuvo: %v", err)
		}
	})
}

// =====================================================================
// SetPersonalityAnswer
// =====================================================================

func TestSetPersonalityAnswer(t *testing.T) {
	t.Run("puntuación válida se acepta", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		a, err := svc.SetPersonalityAnswer(context.Background(), uuid.New(), "extra_funny_laughs", 4)
		if err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
		if a.Score != 4 {
			t.Errorf("score inesperado: %d", a.Score)
		}
	})

	for _, bad := range []int{0, 6, -2} {
		t.Run("puntuación fuera de rango se rechaza", func(t *testing.T) {
			repo := &fakeRepository{}
			svc := newTestService(repo)

			_, err := svc.SetPersonalityAnswer(context.Background(), uuid.New(), "extra_funny_laughs", bad)
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("score %d: se esperaba ValidationError, se obtuvo: %v", bad, err)
			}
			if len(repo.upsertPersonalityCalls) != 0 {
				t.Error("no debería haber llegado a llamar al repositorio")
			}
		})
	}
}

// =====================================================================
// UpdatePartnerPreferences
// =====================================================================

func TestUpdatePartnerPreferences(t *testing.T) {
	t.Run("rango de edad válido se acepta", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		patch := PartnerPreferencesPatch{
			AgeMinSet: true, AgeMin: intPtr(30),
			AgeMaxSet: true, AgeMax: intPtr(50),
		}
		if _, err := svc.UpdatePartnerPreferences(context.Background(), uuid.New(), patch); err != nil {
			t.Errorf("no se esperaba error: %v", err)
		}
	})

	t.Run("age_min mayor que age_max se rechaza sin llegar al repositorio", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		patch := PartnerPreferencesPatch{
			AgeMinSet: true, AgeMin: intPtr(50),
			AgeMaxSet: true, AgeMax: intPtr(30),
		}
		_, err := svc.UpdatePartnerPreferences(context.Background(), uuid.New(), patch)
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("se esperaba ValidationError, se obtuvo: %v", err)
		}
		if len(repo.upsertPartnerCalls) != 0 {
			t.Error("no debería haber llegado a llamar al repositorio")
		}
	})

	t.Run("height_min mayor que height_max se rechaza", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		patch := PartnerPreferencesPatch{
			HeightMinSet: true, HeightMin: intPtr(190),
			HeightMaxSet: true, HeightMax: intPtr(160),
		}
		_, err := svc.UpdatePartnerPreferences(context.Background(), uuid.New(), patch)
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Errorf("se esperaba ValidationError, se obtuvo: %v", err)
		}
	})

	t.Run("solo age_min sin age_max no se puede comparar y se acepta", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		patch := PartnerPreferencesPatch{AgeMinSet: true, AgeMin: intPtr(30)}
		if _, err := svc.UpdatePartnerPreferences(context.Background(), uuid.New(), patch); err != nil {
			t.Errorf("no se esperaba error: %v", err)
		}
	})

	t.Run("about_partner_text demasiado largo se rechaza", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		tooLong := make([]rune, MaxAboutPartnerTextLen+1)
		for i := range tooLong {
			tooLong[i] = 'a'
		}
		s := string(tooLong)
		patch := PartnerPreferencesPatch{AboutPartnerTextSet: true, AboutPartnerText: &s}

		_, err := svc.UpdatePartnerPreferences(context.Background(), uuid.New(), patch)
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Errorf("se esperaba ValidationError, se obtuvo: %v", err)
		}
	})

	t.Run("importancia fuera de rango se rechaza", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		patch := PartnerPreferencesPatch{ImportanceFunSet: true, ImportanceFun: intPtr(9)}
		_, err := svc.UpdatePartnerPreferences(context.Background(), uuid.New(), patch)
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Errorf("se esperaba ValidationError, se obtuvo: %v", err)
		}
	})

	t.Run("importancia dentro de rango en cualquier aspecto se acepta", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		patch := PartnerPreferencesPatch{ImportanceSharedHumorSet: true, ImportanceSharedHumor: intPtr(5)}
		if _, err := svc.UpdatePartnerPreferences(context.Background(), uuid.New(), patch); err != nil {
			t.Errorf("no se esperaba error: %v", err)
		}
	})
}

// =====================================================================
// Validaciones de texto libre nuevas (mismo patrón que TestValidateBio)
// =====================================================================

func TestValidateProfileQuote(t *testing.T) {
	if err := validateProfileQuote(nil); err != nil {
		t.Errorf("profile_quote ausente (nil) no debería ser un error: %v", err)
	}

	short := "Me encanta viajar."
	if err := validateProfileQuote(&short); err != nil {
		t.Errorf("profile_quote corta válida rechazada: %v", err)
	}

	tooLongRunes := make([]rune, MaxProfileQuoteLen+1)
	for i := range tooLongRunes {
		tooLongRunes[i] = 'a'
	}
	tooLong := string(tooLongRunes)
	if err := validateProfileQuote(&tooLong); err == nil {
		t.Error("profile_quote demasiado larga debería rechazarse")
	}
}

func TestValidateDreamWish(t *testing.T) {
	if err := validateDreamWish(nil); err != nil {
		t.Errorf("dream_wish ausente (nil) no debería ser un error: %v", err)
	}

	short := "Visitar todos los rincones del mundo."
	if err := validateDreamWish(&short); err != nil {
		t.Errorf("dream_wish corto válido rechazado: %v", err)
	}

	tooLongRunes := make([]rune, MaxDreamWishLen+1)
	for i := range tooLongRunes {
		tooLongRunes[i] = 'a'
	}
	tooLong := string(tooLongRunes)
	if err := validateDreamWish(&tooLong); err == nil {
		t.Error("dream_wish demasiado largo debería rechazarse")
	}
}
