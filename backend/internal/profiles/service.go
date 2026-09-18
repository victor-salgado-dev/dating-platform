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