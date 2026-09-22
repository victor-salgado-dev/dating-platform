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
// lógica de negocio de Service sin tocar una base de datos real.
// =====================================================================

type fakeRepository struct {
	profile *Profile
	getErr  error

	upsertLanguageCalls []ProfileLanguage
	upsertLanguageErr   error

	interestDefs      map[string]InterestDefinition
	upsertInterestCalls []ProfileInterest
	upsertInterestErr   error

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

// --- Idiomas -------------------------------------------------------------

func (f *fakeRepository) UpsertProfileLanguage(ctx context.Context, profileID uuid.UUID, languageCode string, level *int) (*ProfileLanguage, error) {
	if f.upsertLanguageErr != nil {
		return nil, f.upsertLanguageErr
	}
	pl := ProfileLanguage{ProfileID: profileID, LanguageCode: languageCode, Level: level, UpdatedAt: time.Now()}
	f.upsertLanguageCalls = append(f.upsertLanguageCalls, pl)
	return &pl, nil
}
func (f *fakeRepository) ListProfileLanguages(ctx context.Context, profileID uuid.UUID) ([]ProfileLanguage, error) {
	return nil, nil
}
func (f *fakeRepository) DeleteProfileLanguage(ctx context.Context, profileID uuid.UUID, languageCode string) error {
	return nil
}

// --- Intereses -------------------------------------------------------------

func (f *fakeRepository) ListInterestDefinitions(ctx context.Context) ([]InterestDefinition, error) {
	return nil, nil
}

func (f *fakeRepository) GetInterestDefinition(ctx context.Context, key string) (*InterestDefinition, error) {
	if d, ok := f.interestDefs[key]; ok {
		return &d, nil
	}
	return nil, ErrInterestNotFound
}

func (f *fakeRepository) UpsertProfileInterest(ctx context.Context, profileID uuid.UUID, interestKey string, level *int) (*ProfileInterest, error) {
	if f.upsertInterestErr != nil {
		return nil, f.upsertInterestErr
	}
	pi := ProfileInterest{ProfileID: profileID, InterestKey: interestKey, Level: level, UpdatedAt: time.Now()}
	f.upsertInterestCalls = append(f.upsertInterestCalls, pi)
	return &pi, nil
}
func (f *fakeRepository) ListProfileInterests(ctx context.Context, profileID uuid.UUID) ([]ProfileInterest, error) {
	return nil, nil
}
func (f *fakeRepository) DeleteProfileInterest(ctx context.Context, profileID uuid.UUID, interestKey string) error {
	return nil
}

// --- Personalidad ----------------------------------------------------------

func (f *fakeRepository) ListPersonalityStatements(ctx context.Context) ([]PersonalityStatement, error) {
	return nil, nil
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

// --- Preferencias de pareja --------------------------------------------------

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
// SetLanguage
// =====================================================================

func TestSetLanguage(t *testing.T) {
	t.Run("idioma con nivel válido se acepta", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		pl, err := svc.SetLanguage(context.Background(), uuid.New(), "es", intPtr(5))
		if err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
		if pl.LanguageCode != "es" || pl.Level == nil || *pl.Level != 5 {
			t.Errorf("resultado inesperado: %+v", pl)
		}
	})

	t.Run("idioma sin nivel se acepta", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(repo)

		_, err := svc.SetLanguage(context.Background(), uuid.New(), "en", nil)
		if err != nil {
			t.Errorf("no se esperaba error: %v", err)
		}
	})

	for _, bad := range []int{0, 6, -1} {
		t.Run("nivel fuera de rango se rechaza antes de llegar al repositorio", func(t *testing.T) {
			repo := &fakeRepository{}
			svc := newTestService(repo)

			_, err := svc.SetLanguage(context.Background(), uuid.New(), "fr", intPtr(bad))
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("nivel %d: se esperaba ValidationError, se obtuvo: %v", bad, err)
			}
			if len(repo.upsertLanguageCalls) != 0 {
				t.Error("no debería haber llegado a llamar al repositorio")
			}
		})
	}

	t.Run("código de idioma inválido lo rechaza la BD, no el Service", func(t *testing.T) {
		// El Service no valida la lista de códigos (vive como CHECK en
		// la BD); aquí solo comprobamos que el error del repositorio se
		// propaga tal cual.
		repo := &fakeRepository{upsertLanguageErr: invalidField("language", "código no reconocido")}
		svc := newTestService(repo)

		_, err := svc.SetLanguage(context.Background(), uuid.New(), "klingon", intPtr(3))
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Errorf("se esperaba que el ValidationError del repositorio se propagara, se obtuvo: %v", err)
		}
	})
}

// =====================================================================
// SetInterest — la regla has_level <-> level
// =====================================================================

func TestSetInterest(t *testing.T) {
	leveled := InterestDefinition{Key: "cocina", HasLevel: true}
	concrete := InterestDefinition{Key: "guitarra", HasLevel: false}

	t.Run("interés con nivel, con nivel dado, se acepta", func(t *testing.T) {
		repo := &fakeRepository{interestDefs: map[string]InterestDefinition{"cocina": leveled}}
		svc := newTestService(repo)

		pi, err := svc.SetInterest(context.Background(), uuid.New(), "cocina", intPtr(4))
		if err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
		if pi.Level == nil || *pi.Level != 4 {
			t.Errorf("resultado inesperado: %+v", pi)
		}
	})

	t.Run("interés con nivel, SIN nivel, se rechaza", func(t *testing.T) {
		repo := &fakeRepository{interestDefs: map[string]InterestDefinition{"cocina": leveled}}
		svc := newTestService(repo)

		_, err := svc.SetInterest(context.Background(), uuid.New(), "cocina", nil)
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("se esperaba ValidationError, se obtuvo: %v", err)
		}
		if len(repo.upsertInterestCalls) != 0 {
			t.Error("no debería haber llegado a llamar al repositorio")
		}
	})

	t.Run("interés concreto, sin nivel, se acepta", func(t *testing.T) {
		repo := &fakeRepository{interestDefs: map[string]InterestDefinition{"guitarra": concrete}}
		svc := newTestService(repo)

		pi, err := svc.SetInterest(context.Background(), uuid.New(), "guitarra", nil)
		if err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
		if pi.Level != nil {
			t.Errorf("un interés concreto no debería guardar nivel, se obtuvo: %v", *pi.Level)
		}
	})

	t.Run("interés concreto, CON nivel, se rechaza", func(t *testing.T) {
		repo := &fakeRepository{interestDefs: map[string]InterestDefinition{"guitarra": concrete}}
		svc := newTestService(repo)

		_, err := svc.SetInterest(context.Background(), uuid.New(), "guitarra", intPtr(3))
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("se esperaba ValidationError, se obtuvo: %v", err)
		}
		if len(repo.upsertInterestCalls) != 0 {
			t.Error("no debería haber llegado a llamar al repositorio")
		}
	})

	for _, bad := range []int{0, 6, -1} {
		t.Run("nivel fuera de rango 1-5 se rechaza", func(t *testing.T) {
			repo := &fakeRepository{interestDefs: map[string]InterestDefinition{"cocina": leveled}}
			svc := newTestService(repo)

			_, err := svc.SetInterest(context.Background(), uuid.New(), "cocina", intPtr(bad))
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("nivel %d: se esperaba ValidationError, se obtuvo: %v", bad, err)
			}
		})
	}

	t.Run("interest_key inexistente en el catálogo se rechaza con un mensaje claro", func(t *testing.T) {
		repo := &fakeRepository{interestDefs: map[string]InterestDefinition{}}
		svc := newTestService(repo)

		_, err := svc.SetInterest(context.Background(), uuid.New(), "no_existe", intPtr(3))
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("se esperaba ValidationError, se obtuvo: %v", err)
		}
		if valErr.Field != "interest_key" {
			t.Errorf("se esperaba el campo 'interest_key', fue %q", valErr.Field)
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
}
