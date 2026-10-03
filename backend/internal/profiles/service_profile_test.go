package profiles

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// fakeRepo implementa solo lo que usan estos tests. Embebe la interfaz para
// satisfacerla: cualquier otro método entra en pánico (nil), lo que delata un
// uso inesperado del repositorio.
type fakeRepo struct {
	Repository

	created     *Profile
	updatedWith *ProfilePatch
	prefsWith   *PartnerPreferencesPatch

	myProfileID uuid.UUID // perfil que devuelve GetByUserID (aleatorio si es Nil)
	noProfile   bool      // GetByUserID devuelve ErrNotFound
}

func (f *fakeRepo) Create(_ context.Context, p *Profile) error {
	p.ID = uuid.New()
	f.created = p
	return nil
}

func (f *fakeRepo) Update(_ context.Context, _ uuid.UUID, patch ProfilePatch) (*Profile, error) {
	f.updatedWith = &patch
	return &Profile{}, nil
}

func (f *fakeRepo) GetByUserID(context.Context, uuid.UUID) (*Profile, error) {
	if f.noProfile {
		return nil, ErrNotFound
	}
	id := f.myProfileID
	if id == uuid.Nil {
		id = uuid.New()
	}
	return &Profile{ID: id}, nil
}

// Lecturas que hace GetFullPublicProfile.
func (f *fakeRepo) GetPublicByID(_ context.Context, id, _ uuid.UUID) (*Profile, error) {
	return &Profile{ID: id}, nil
}
func (f *fakeRepo) ListPhotos(context.Context, uuid.UUID) ([]Photo, error) { return nil, nil }
func (f *fakeRepo) ListProfileLanguages(context.Context, uuid.UUID) ([]ProfileLanguage, error) {
	return nil, nil
}
func (f *fakeRepo) ListProfileInterests(context.Context, uuid.UUID) ([]ProfileInterest, error) {
	return nil, nil
}
func (f *fakeRepo) ListInterestDefinitions(context.Context) ([]InterestDefinition, error) {
	return nil, nil
}
func (f *fakeRepo) ListPersonalityAnswers(context.Context, uuid.UUID) ([]ProfilePersonalityAnswer, error) {
	return nil, nil
}
func (f *fakeRepo) GetPersonalityTraitScores(context.Context, uuid.UUID) ([]PersonalityTraitScore, error) {
	return nil, nil
}
func (f *fakeRepo) GetPartnerPreferences(_ context.Context, id uuid.UUID) (*PartnerPreferences, error) {
	return &PartnerPreferences{ProfileID: id}, nil
}

func (f *fakeRepo) UpsertPartnerPreferences(_ context.Context, _ uuid.UUID, patch PartnerPreferencesPatch) (*PartnerPreferences, error) {
	f.prefsWith = &patch
	return &PartnerPreferences{}, nil
}

func validProfile() *Profile {
	nat := " de "
	return &Profile{
		DisplayName: "  Ana  ",
		BirthDate:   time.Date(1990, 5, 17, 0, 0, 0, 0, time.UTC),
		Gender:      GenderFemale,
		CountryCode: " es ",
		ProfileDetails: ProfileDetails{
			Nationality:       &nat,
			RelationshipGoals: []RelationshipGoal{RelationshipLongTerm},
		},
	}
}

func TestCreateProfile_NormalizesWithoutMutatingInput(t *testing.T) {
	repo := &fakeRepo{}
	svc := &Service{repo: repo}
	in := validProfile()
	userID := uuid.New()

	out, err := svc.CreateProfile(context.Background(), userID, in)
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if out.UserID != userID || out.ID == uuid.Nil {
		t.Errorf("faltan UserID/ID en el resultado: %+v", out)
	}
	if out.DisplayName != "Ana" || out.CountryCode != "ES" || out.Nationality == nil || *out.Nationality != "DE" {
		t.Errorf("normalización incorrecta: name=%q country=%q nationality=%v", out.DisplayName, out.CountryCode, out.Nationality)
	}
	if repo.created != out {
		t.Error("el perfil devuelto debe ser el que se guardó")
	}
	// El llamador no ve su struct modificado.
	if in.DisplayName != "  Ana  " || in.CountryCode != " es " || *in.Nationality != " de " || in.UserID != uuid.Nil {
		t.Errorf("CreateProfile modificó el perfil de entrada: %+v", in)
	}
}

func TestCreateProfile_Validation(t *testing.T) {
	long := strings.Repeat("x", MaxBioLen+1)
	cases := []struct {
		name   string
		mutate func(p *Profile)
	}{
		{"nombre vacío", func(p *Profile) { p.DisplayName = "   " }},
		{"menor de edad", func(p *Profile) { p.BirthDate = time.Now().AddDate(-10, 0, 0) }},
		{"sin fecha", func(p *Profile) { p.BirthDate = time.Time{} }},
		{"género no permitido", func(p *Profile) { p.Gender = "robot" }},
		{"país inválido", func(p *Profile) { p.CountryCode = "ESP" }},
		{"objetivo no permitido", func(p *Profile) { p.RelationshipGoals = []RelationshipGoal{"nada"} }},
		{"bio demasiado larga", func(p *Profile) { p.Bio = &long }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{}
			p := validProfile()
			tc.mutate(p)
			if _, err := (&Service{repo: repo}).CreateProfile(context.Background(), uuid.New(), p); err == nil {
				t.Fatal("se esperaba un error de validación")
			}
			if repo.created != nil {
				t.Error("no debe guardarse un perfil inválido")
			}
		})
	}
}

func TestUpdateProfile_EmptyPatchPassesThrough(t *testing.T) {
	repo := &fakeRepo{}
	if _, err := (&Service{repo: repo}).UpdateProfile(context.Background(), uuid.New(), ProfilePatch{}); err != nil {
		t.Fatalf("un patch vacío no debe fallar: %v", err)
	}
	if repo.updatedWith == nil {
		t.Fatal("debe llegar al repositorio")
	}
	if cols, _ := setColumns(*repo.updatedWith); len(cols) != 0 {
		t.Errorf("no debía haber columnas, hay %v", cols)
	}
}

func TestUpdateProfile_ValidatesOnlySetFields(t *testing.T) {
	long := strings.Repeat("x", MaxBioLen+1)
	cases := []struct {
		name  string
		patch ProfilePatch
	}{
		{"nombre vacío", ProfilePatch{DisplayName: FieldOf("  ")}},
		{"nombre null (llega como vacío)", ProfilePatch{DisplayName: FieldOf("")}},
		{"menor de edad", ProfilePatch{BirthDate: FieldOf(DateOnly(time.Now().AddDate(-10, 0, 0)))}},
		{"fecha null (llega como cero)", ProfilePatch{BirthDate: FieldOf(DateOnly{})}},
		{"género no permitido", ProfilePatch{Gender: FieldOf(Gender("robot"))}},
		{"país inválido", ProfilePatch{CountryCode: FieldOf("ESP")}},
		{"objetivo no permitido", ProfilePatch{RelationshipGoals: FieldOf([]RelationshipGoal{"nada"})}},
		{"bio demasiado larga", ProfilePatch{Bio: FieldOf(&long)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{}
			if _, err := (&Service{repo: repo}).UpdateProfile(context.Background(), uuid.New(), tc.patch); err == nil {
				t.Fatal("se esperaba un error de validación")
			}
			if repo.updatedWith != nil {
				t.Error("no debe llegar al repositorio un patch inválido")
			}
		})
	}

	// Borrar (null) un opcional es válido.
	repo := &fakeRepo{}
	ok := ProfilePatch{Bio: FieldOf[*string](nil), Region: FieldOf[*string](nil)}
	if _, err := (&Service{repo: repo}).UpdateProfile(context.Background(), uuid.New(), ok); err != nil {
		t.Errorf("borrar campos opcionales debe ser válido: %v", err)
	}
}

func TestUpdateProfile_Normalizes(t *testing.T) {
	repo := &fakeRepo{}
	nat := " de "
	patch := ProfilePatch{
		DisplayName: FieldOf("  Ana "),
		CountryCode: FieldOf(" es "),
		Nationality: FieldOf(&nat),
	}

	if _, err := (&Service{repo: repo}).UpdateProfile(context.Background(), uuid.New(), patch); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}

	got := repo.updatedWith
	if got.DisplayName.Value != "Ana" || got.CountryCode.Value != "ES" || *got.Nationality.Value != "DE" {
		t.Errorf("normalización incorrecta: %q %q %q", got.DisplayName.Value, got.CountryCode.Value, *got.Nationality.Value)
	}
	if nat != " de " {
		t.Errorf("no debe modificarse el valor al que apunta el patch del llamador: %q", nat)
	}
}

func TestUpdatePartnerPreferences_Validation(t *testing.T) {
	i := func(n int) *int { return &n }
	long := strings.Repeat("x", MaxAboutPartnerTextLen+1)

	bad := map[string]PartnerPreferencesPatch{
		"edad mínima mayor que máxima":   {AgeMin: FieldOf(i(50)), AgeMax: FieldOf(i(40))},
		"altura mínima mayor que máxima": {HeightMin: FieldOf(i(190)), HeightMax: FieldOf(i(160))},
		"texto demasiado largo":          {AboutPartnerText: FieldOf(&long)},
		"importancia por debajo":         {ImportanceFun: FieldOf(i(MinPartnerImportance - 1))},
		"importancia por encima":         {ImportanceIndependence: FieldOf(i(MaxPartnerImportance + 1))},
	}
	for name, patch := range bad {
		t.Run(name, func(t *testing.T) {
			repo := &fakeRepo{}
			if _, err := (&Service{repo: repo}).UpdatePartnerPreferences(context.Background(), uuid.New(), patch); err == nil {
				t.Fatal("se esperaba un error de validación")
			}
			if repo.prefsWith != nil {
				t.Error("no debe llegar al repositorio")
			}
		})
	}

	// Un patch válido (incluido borrar con null) llega al repositorio.
	repo := &fakeRepo{}
	good := PartnerPreferencesPatch{
		AgeMin:        FieldOf(i(30)),
		AgeMax:        FieldOf(i(40)),
		ImportanceFun: FieldOf[*int](nil),
	}
	if _, err := (&Service{repo: repo}).UpdatePartnerPreferences(context.Background(), uuid.New(), good); err != nil {
		t.Fatalf("patch válido rechazado: %v", err)
	}
	if repo.prefsWith == nil || !repo.prefsWith.ImportanceFun.Set {
		t.Errorf("el patch no llegó al repositorio intacto: %+v", repo.prefsWith)
	}
}
