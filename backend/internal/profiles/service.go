package profiles

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/storage"
)

const (
	MinAge              = 18 // V1 es solo para mayores de edad (sección 1).
	MaxDisplayNameLen   = 100
	MaxBioLen           = 1000
	MaxInterests        = 20
	MaxInterestLen      = 40
	MaxLanguages        = 10
	MaxPhotosPerProfile = 6
	MaxPhotoSizeBytes   = 5 * 1024 * 1024 // 5 MB

	MaxProfileQuoteLen      = 500
	MaxDreamWishLen         = 500
	MaxAboutPartnerTextLen  = 1000
	MinPersonalityScore     = 1
	MaxPersonalityScore     = 5
	MinHobbyIntensity       = 1
	MaxHobbyIntensity       = 5
	MinPartnerImportance    = 1
	MaxPartnerImportance    = 5
)

var countryCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)

var allowedGenders = map[Gender]bool{
	GenderFemale: true, GenderMale: true, GenderNonBinary: true, GenderOther: true,
}

var allowedRelationshipGoals = map[RelationshipGoal]bool{
	RelationshipCasual: true, RelationshipLongTerm: true, RelationshipFriendship: true,
	RelationshipMarriage: true, RelationshipNotSure: true,
}

func IsValidGender(g Gender) bool {
	return allowedGenders[g]
}

func IsValidRelationshipGoal(g RelationshipGoal) bool {
	return allowedRelationshipGoals[g]
}

var allowedPersonalityTraits = map[PersonalityTrait]bool{
	TraitExtraversion: true, TraitEmotionalStability: true, TraitConscientiousness: true,
	TraitAgreeableness: true, TraitOpenness: true,
}

// IsValidPersonalityTrait existe por el mismo motivo que IsValidGender:
// el paquete search (Fase 5) necesita validar un trait_key recibido por
// query string antes de usarlo en una consulta, y los 5 rasgos son un
// catálogo cerrado (a diferencia de las afirmaciones dentro de cada
// rasgo, que sí viven en una tabla porque esas sí crecen).
func IsValidPersonalityTrait(t PersonalityTrait) bool {
	return allowedPersonalityTraits[t]
}

var allowedPhotoTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// CreateProfileInput son los datos necesarios para crear un perfil.
type CreateProfileInput struct {
	DisplayName      string
	BirthDate        time.Time
	Gender           Gender
	CountryCode      string
	Region           *string
	Languages        []string
	RelationshipGoal *RelationshipGoal
	HasChildren      *string
	WantsChildren    *string
	Bio              *string
	Interests        []string

	// --- Nuevos campos ---
	Height                *int
	Weight                *int
	BodyType              *string
	Ethnicity             *string
	AppearanceRating      *string
	HairColor             *string
	EyeColor              *string
	BodyArt               []string
	SmokingHabit          *string
	DrinkingHabit         *string
	RelocationWillingness []string
	MaritalStatus         *string
	ChildrenCount         *int
	YoungestChildAge      *int
	OldestChildAge        *int
	Occupation            *string
	EmploymentStatus      *string
	IncomeLevel           *string
	LivingSituation       *string
	Nationality           *string
	EducationLevel        *string
	EnglishAbility        *string
	Religion              *string
	ReligiousValues       *string
	StarSign              *string

	// --- NUEVOS CAMPOS (Über mich / estilo de vida) ---
	FutureVision       []string
	Sports             []string
	LikesPets          *string
	PetsOwned          []string
	FavoriteSeason     *string
	IdealVacationStyle []string
	VacationActivities []string
	ProfileQuote       *string
	DreamWish          *string
}

type Service struct {
	repo    Repository
	storage storage.Storage
}

func NewService(repo Repository, store storage.Storage) *Service {
	return &Service{repo: repo, storage: store}
}

func (s *Service) GetMyProfile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) GetPublicProfile(ctx context.Context, viewerUserID, profileID uuid.UUID) (*Profile, error) {
	return s.repo.GetPublicByID(ctx, profileID, viewerUserID)
}

func (s *Service) ListPublicPhotos(ctx context.Context, viewerUserID, profileID uuid.UUID) ([]Photo, error) {
	if _, err := s.repo.GetPublicByID(ctx, profileID, viewerUserID); err != nil {
		return nil, err
	}
	return s.repo.ListPhotos(ctx, profileID)
}

func (s *Service) OpenPublicPhoto(ctx context.Context, viewerUserID, profileID, photoID uuid.UUID) (io.ReadCloser, *Photo, error) {
	if _, err := s.repo.GetPublicByID(ctx, profileID, viewerUserID); err != nil {
		return nil, nil, err
	}

	ph, err := s.repo.GetPhoto(ctx, profileID, photoID)
	if err != nil {
		return nil, nil, err
	}

	rc, err := s.storage.Open(ctx, ph.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("profiles: abrir fichero de foto: %w", err)
	}

	return rc, ph, nil
}

func (s *Service) CreateProfile(ctx context.Context, userID uuid.UUID, in CreateProfileInput) (*Profile, error) {
	if err := validateDisplayName(in.DisplayName); err != nil {
		return nil, err
	}
	if err := validateBirthDate(in.BirthDate); err != nil {
		return nil, err
	}
	if err := validateGender(in.Gender); err != nil {
		return nil, err
	}
	countryCode := strings.ToUpper(strings.TrimSpace(in.CountryCode))
	if err := validateCountryCode(countryCode); err != nil {
		return nil, err
	}
	if err := validateOptionalRelationshipGoal(in.RelationshipGoal); err != nil {
		return nil, err
	}
	if err := validateBio(in.Bio); err != nil {
		return nil, err
	}
	if err := validateInterests(in.Interests); err != nil {
		return nil, err
	}
	if err := validateLanguages(in.Languages); err != nil {
		return nil, err
	}
	if err := validateProfileQuote(in.ProfileQuote); err != nil {
		return nil, err
	}
	if err := validateDreamWish(in.DreamWish); err != nil {
		return nil, err
	}

	var nationality *string
	if in.Nationality != nil {
		n := strings.ToUpper(strings.TrimSpace(*in.Nationality))
		nationality = &n
	}

	p := &Profile{
		UserID:           userID,
		DisplayName:      strings.TrimSpace(in.DisplayName),
		BirthDate:        in.BirthDate,
		Gender:           in.Gender,
		CountryCode:      countryCode,
		Region:           in.Region,
		Languages:        in.Languages,
		RelationshipGoal: in.RelationshipGoal,
		HasChildren:      in.HasChildren,
		WantsChildren:    in.WantsChildren,
		Bio:              in.Bio,
		Interests:        in.Interests,

		// Asignación de los nuevos campos
		Height:                in.Height,
		Weight:                in.Weight,
		BodyType:              in.BodyType,
		Ethnicity:             in.Ethnicity,
		AppearanceRating:      in.AppearanceRating,
		HairColor:             in.HairColor,
		EyeColor:              in.EyeColor,
		BodyArt:               in.BodyArt,
		SmokingHabit:          in.SmokingHabit,
		DrinkingHabit:         in.DrinkingHabit,
		RelocationWillingness: in.RelocationWillingness,
		MaritalStatus:         in.MaritalStatus,
		ChildrenCount:         in.ChildrenCount,
		YoungestChildAge:      in.YoungestChildAge,
		OldestChildAge:        in.OldestChildAge,
		Occupation:            in.Occupation,
		EmploymentStatus:      in.EmploymentStatus,
		IncomeLevel:           in.IncomeLevel,
		LivingSituation:       in.LivingSituation,
		Nationality:           nationality,
		EducationLevel:        in.EducationLevel,
		EnglishAbility:        in.EnglishAbility,
		Religion:              in.Religion,
		ReligiousValues:       in.ReligiousValues,
		StarSign:              in.StarSign,

		// Über mich / estilo de vida
		FutureVision:       in.FutureVision,
		Sports:             in.Sports,
		LikesPets:          in.LikesPets,
		PetsOwned:          in.PetsOwned,
		FavoriteSeason:     in.FavoriteSeason,
		IdealVacationStyle: in.IdealVacationStyle,
		VacationActivities: in.VacationActivities,
		ProfileQuote:       in.ProfileQuote,
		DreamWish:          in.DreamWish,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, patch ProfilePatch) (*Profile, error) {
	if patch.DisplayName != nil {
		if err := validateDisplayName(*patch.DisplayName); err != nil {
			return nil, err
		}
		trimmed := strings.TrimSpace(*patch.DisplayName)
		patch.DisplayName = &trimmed
	}
	if patch.BirthDate != nil {
		if err := validateBirthDate(*patch.BirthDate); err != nil {
			return nil, err
		}
	}
	if patch.Gender != nil {
		if err := validateGender(*patch.Gender); err != nil {
			return nil, err
		}
	}
	if patch.CountryCode != nil {
		cc := strings.ToUpper(strings.TrimSpace(*patch.CountryCode))
		if err := validateCountryCode(cc); err != nil {
			return nil, err
		}
		patch.CountryCode = &cc
	}
	if patch.Nationality != nil {
		nat := strings.ToUpper(strings.TrimSpace(*patch.Nationality))
		patch.Nationality = &nat
	}
	if patch.RelationshipGoalSet {
		if err := validateOptionalRelationshipGoal(patch.RelationshipGoal); err != nil {
			return nil, err
		}
	}
	if patch.BioSet {
		if err := validateBio(patch.Bio); err != nil {
			return nil, err
		}
	}
	if patch.InterestsSet {
		if err := validateInterests(patch.Interests); err != nil {
			return nil, err
		}
	}
	if patch.LanguagesSet {
		if err := validateLanguages(patch.Languages); err != nil {
			return nil, err
		}
	}
	if patch.ProfileQuoteSet {
		if err := validateProfileQuote(patch.ProfileQuote); err != nil {
			return nil, err
		}
	}
	if patch.DreamWishSet {
		if err := validateDreamWish(patch.DreamWish); err != nil {
			return nil, err
		}
	}

	return s.repo.Update(ctx, userID, patch)
}

func (s *Service) UploadPhoto(ctx context.Context, userID uuid.UUID, declaredContentType string, size int64, r io.Reader) (*Photo, error) {
	if size <= 0 || size > MaxPhotoSizeBytes {
		return nil, invalidField("photo", fmt.Sprintf("el fichero debe pesar menos de %d MB", MaxPhotoSizeBytes/1024/1024))
	}

	peek := make([]byte, 512)
	n, err := io.ReadFull(r, peek)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, fmt.Errorf("profiles: leer foto: %w", err)
	}
	peek = peek[:n]

	detectedType := http.DetectContentType(peek)
	ext, ok := allowedPhotoTypes[detectedType]
	if !ok {
		return nil, invalidField("photo", "el contenido del fichero no es una imagen JPEG, PNG o WebP válida")
	}
	if declaredContentType != "" && declaredContentType != detectedType {
		slog.Warn("content-type declarado no coincide con el detectado",
			"declared", declaredContentType, "detected", detectedType)
	}

	fullReader := io.MultiReader(bytes.NewReader(peek), r)

	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	count, err := s.repo.CountPhotos(ctx, profile.ID)
	if err != nil {
		return nil, err
	}
	if count >= MaxPhotosPerProfile {
		return nil, ErrTooManyPhotos
	}

	key := fmt.Sprintf("profiles/%s/%s%s", profile.ID, uuid.NewString(), ext)
	if err := s.storage.Save(ctx, key, fullReader); err != nil {
		return nil, fmt.Errorf("profiles: guardar fichero de foto: %w", err)
	}

	photo := &Photo{StorageKey: key, ContentType: detectedType}
	if err := s.repo.AddPhoto(ctx, profile.ID, photo); err != nil {
		if delErr := s.storage.Delete(ctx, key); delErr != nil {
			slog.Error("no se pudo limpiar el fichero huérfano tras un fallo", "key", key, "error", delErr)
		}
		return nil, err
	}

	return photo, nil
}

func (s *Service) ListPhotos(ctx context.Context, userID uuid.UUID) ([]Photo, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListPhotos(ctx, profile.ID)
}

func (s *Service) OpenPhoto(ctx context.Context, userID, photoID uuid.UUID) (io.ReadCloser, *Photo, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	ph, err := s.repo.GetPhoto(ctx, profile.ID, photoID)
	if err != nil {
		return nil, nil, err
	}

	rc, err := s.storage.Open(ctx, ph.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("profiles: abrir fichero de foto: %w", err)
	}

	return rc, ph, nil
}

func (s *Service) DeletePhoto(ctx context.Context, userID, photoID uuid.UUID) error {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	ph, err := s.repo.GetPhoto(ctx, profile.ID, photoID)
	if err != nil {
		return err
	}

	if err := s.repo.DeletePhoto(ctx, profile.ID, photoID); err != nil {
		return err
	}

	if err := s.storage.Delete(ctx, ph.StorageKey); err != nil {
		slog.Error("no se pudo borrar el fichero de la foto", "key", ph.StorageKey, "error", err)
	}

	return nil
}

// --- Catálogos --------------------------------------------------------
//
// Son de solo lectura para el usuario final: no hay validación de
// negocio que hacer aquí, el Service simplemente delega en el
// Repository. Existen como métodos propios (en vez de exponer el
// Repository directamente al Handler) para no filtrar el detalle de
// implementación de que el catálogo vive en Postgres.

func (s *Service) ListHobbyCatalog(ctx context.Context) ([]HobbyDefinition, error) {
	return s.repo.ListHobbyDefinitions(ctx)
}

func (s *Service) ListPersonalityCatalog(ctx context.Context) ([]PersonalityStatement, error) {
	return s.repo.ListPersonalityStatements(ctx)
}

// --- Hobbies del usuario -------------------------------------------------

// SetHobby crea o actualiza la respuesta del usuario a un hobby del
// catálogo. hobbyKey inexistente en el catálogo se traduce en un
// ValidationError por parte del Repository (violación de FK).
func (s *Service) SetHobby(ctx context.Context, userID uuid.UUID, hobbyKey string, liked bool, intensity *int) (*ProfileHobby, error) {
	if !liked && intensity != nil {
		return nil, invalidField("intensity", "no tiene sentido indicar una intensidad si no te gusta ese hobby")
	}
	if intensity != nil && (*intensity < MinHobbyIntensity || *intensity > MaxHobbyIntensity) {
		return nil, invalidField("intensity", fmt.Sprintf("debe estar entre %d y %d", MinHobbyIntensity, MaxHobbyIntensity))
	}

	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.UpsertProfileHobby(ctx, profile.ID, hobbyKey, liked, intensity)
}

func (s *Service) ListMyHobbies(ctx context.Context, userID uuid.UUID) ([]ProfileHobby, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListProfileHobbies(ctx, profile.ID)
}

func (s *Service) ListPublicHobbies(ctx context.Context, viewerUserID, profileID uuid.UUID) ([]ProfileHobby, error) {
	if _, err := s.repo.GetPublicByID(ctx, profileID, viewerUserID); err != nil {
		return nil, err
	}
	return s.repo.ListProfileHobbies(ctx, profileID)
}

// DeleteHobby vuelve un hobby a "no contestado". Operación idempotente.
func (s *Service) DeleteHobby(ctx context.Context, userID uuid.UUID, hobbyKey string) error {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.repo.DeleteProfileHobby(ctx, profile.ID, hobbyKey)
}

// --- Personalidad del usuario ---------------------------------------------

func (s *Service) SetPersonalityAnswer(ctx context.Context, userID uuid.UUID, statementKey string, score int) (*ProfilePersonalityAnswer, error) {
	if score < MinPersonalityScore || score > MaxPersonalityScore {
		return nil, invalidField("score", fmt.Sprintf("debe estar entre %d y %d", MinPersonalityScore, MaxPersonalityScore))
	}

	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.UpsertPersonalityAnswer(ctx, profile.ID, statementKey, score)
}

// GetMyPersonality devuelve tanto las respuestas individuales como el
// agregado ("Gesamt") por rasgo, calculado por la BD.
func (s *Service) GetMyPersonality(ctx context.Context, userID uuid.UUID) ([]ProfilePersonalityAnswer, []PersonalityTraitScore, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return s.getPersonality(ctx, profile.ID)
}

func (s *Service) GetPublicPersonality(ctx context.Context, viewerUserID, profileID uuid.UUID) ([]ProfilePersonalityAnswer, []PersonalityTraitScore, error) {
	if _, err := s.repo.GetPublicByID(ctx, profileID, viewerUserID); err != nil {
		return nil, nil, err
	}
	return s.getPersonality(ctx, profileID)
}

func (s *Service) getPersonality(ctx context.Context, profileID uuid.UUID) ([]ProfilePersonalityAnswer, []PersonalityTraitScore, error) {
	answers, err := s.repo.ListPersonalityAnswers(ctx, profileID)
	if err != nil {
		return nil, nil, err
	}
	scores, err := s.repo.GetPersonalityTraitScores(ctx, profileID)
	if err != nil {
		return nil, nil, err
	}
	return answers, scores, nil
}

// --- Preferencias de pareja del usuario -------------------------------------

func (s *Service) GetMyPartnerPreferences(ctx context.Context, userID uuid.UUID) (*PartnerPreferences, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetPartnerPreferences(ctx, profile.ID)
}

func (s *Service) UpdatePartnerPreferences(ctx context.Context, userID uuid.UUID, patch PartnerPreferencesPatch) (*PartnerPreferences, error) {
	if patch.AgeMinSet && patch.AgeMaxSet && patch.AgeMin != nil && patch.AgeMax != nil && *patch.AgeMin > *patch.AgeMax {
		return nil, invalidField("age_min", "no puede ser mayor que age_max")
	}
	if patch.HeightMinSet && patch.HeightMaxSet && patch.HeightMin != nil && patch.HeightMax != nil && *patch.HeightMin > *patch.HeightMax {
		return nil, invalidField("height_min", "no puede ser mayor que height_max")
	}
	if patch.AboutPartnerTextSet && patch.AboutPartnerText != nil && len([]rune(*patch.AboutPartnerText)) > MaxAboutPartnerTextLen {
		return nil, invalidField("about_partner_text", fmt.Sprintf("no puede superar %d caracteres", MaxAboutPartnerTextLen))
	}
	for _, imp := range []*int{
		patch.ImportanceSharedThoughts, patch.ImportanceSharedHobbies, patch.ImportanceIntimacy,
		patch.ImportanceRomanticLove, patch.ImportanceFinancialSecurity, patch.ImportanceFun,
		patch.ImportanceSharedFriends, patch.ImportanceSharedHumor, patch.ImportancePersonalSpace,
		patch.ImportanceIndependence,
	} {
		if imp != nil && (*imp < MinPartnerImportance || *imp > MaxPartnerImportance) {
			return nil, invalidField("importance", fmt.Sprintf("debe estar entre %d y %d", MinPartnerImportance, MaxPartnerImportance))
		}
	}

	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.UpsertPartnerPreferences(ctx, profile.ID, patch)
}

// --- Validaciones auxiliares ---

func validateDisplayName(v string) error {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return invalidField("display_name", "es obligatorio")
	}
	if len([]rune(trimmed)) > MaxDisplayNameLen {
		return invalidField("display_name", fmt.Sprintf("no puede superar %d caracteres", MaxDisplayNameLen))
	}
	return nil
}

func validateBirthDate(t time.Time) error {
	if t.IsZero() {
		return invalidField("birth_date", "es obligatorio")
	}
	if AgeAt(t, time.Now()) < MinAge {
		return invalidField("birth_date", "debes tener al menos 18 años para usar la plataforma")
	}
	return nil
}

func validateGender(g Gender) error {
	if !allowedGenders[g] {
		return invalidField("gender", "valor no permitido")
	}
	return nil
}

func validateCountryCode(v string) error {
	if !countryCodePattern.MatchString(v) {
		return invalidField("country_code", "debe ser un código ISO 3166-1 alpha-2 (p.ej. ES, FR, US)")
	}
	return nil
}

func validateOptionalRelationshipGoal(g *RelationshipGoal) error {
	if g == nil {
		return nil
	}
	if !allowedRelationshipGoals[*g] {
		return invalidField("relationship_goal", "valor no permitido")
	}
	return nil
}

func validateBio(v *string) error {
	if v == nil {
		return nil
	}
	if len([]rune(*v)) > MaxBioLen {
		return invalidField("bio", fmt.Sprintf("no puede superar %d caracteres", MaxBioLen))
	}
	return nil
}

func validateInterests(v []string) error {
	if len(v) > MaxInterests {
		return invalidField("interests", fmt.Sprintf("no puedes indicar más de %d intereses", MaxInterests))
	}
	for _, i := range v {
		if strings.TrimSpace(i) == "" {
			return invalidField("interests", "no puede contener valores vacíos")
		}
		if len([]rune(i)) > MaxInterestLen {
			return invalidField("interests", fmt.Sprintf("cada interés debe tener menos de %d caracteres", MaxInterestLen))
		}
	}
	return nil
}

func validateLanguages(v []string) error {
	if len(v) > MaxLanguages {
		return invalidField("languages", fmt.Sprintf("no puedes indicar más de %d idiomas", MaxLanguages))
	}
	for _, l := range v {
		if strings.TrimSpace(l) == "" {
			return invalidField("languages", "no puede contener valores vacíos")
		}
	}
	return nil
}

func validateProfileQuote(v *string) error {
	if v == nil {
		return nil
	}
	if len([]rune(*v)) > MaxProfileQuoteLen {
		return invalidField("profile_quote", fmt.Sprintf("no puede superar %d caracteres", MaxProfileQuoteLen))
	}
	return nil
}

func validateDreamWish(v *string) error {
	if v == nil {
		return nil
	}
	if len([]rune(*v)) > MaxDreamWishLen {
		return invalidField("dream_wish", fmt.Sprintf("no puede superar %d caracteres", MaxDreamWishLen))
	}
	return nil
}
