package profiles

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/storage"
)

const (
	MinAge              = 18 // V1 es solo para mayores de edad (sección 1).
	MaxDisplayNameLen   = 100
	MaxBioLen           = 1000
	MaxPhotosPerProfile = 6
	MaxPhotoSizeBytes   = 5 * 1024 * 1024 // 5 MB

	MaxProfileQuoteLen     = 500
	MaxDreamWishLen        = 500
	MaxAboutPartnerTextLen = 1000

	MinPersonalityScore = 1
	MaxPersonalityScore = 5

	MinInterestLevel = 1
	MaxInterestLevel = 5

	MinLanguageLevel = 1
	MaxLanguageLevel = 5

	MinPartnerImportance = 1
	MaxPartnerImportance = 5
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
// el paquete search necesita validar un trait_key recibido por query
// string antes de usarlo en una consulta, y los 5 rasgos son un
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

// FavoriteChecker resume la única operación de favoritos que necesita
// el servicio de perfiles para calcular el estado del visitante.
type FavoriteChecker interface {
	IsFavorited(ctx context.Context, userID, profileID uuid.UUID) (bool, error)
}

// LikeChecker resume las operaciones de likes/matches necesarias para
// calcular el estado del visitante.
type LikeChecker interface {
	IsLiked(ctx context.Context, fromProfileID, toProfileID uuid.UUID) (bool, error)
	HasMatch(ctx context.Context, profileA, profileB uuid.UUID) (bool, error)
}

// BlockChecker resume la operación de bloqueo necesaria para saber si
// existe bloqueo en cualquier sentido entre dos usuarios.
type BlockChecker interface {
	IsBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error)
}

// VisitRecorder permite registrar una visita a un perfil.
type VisitRecorder interface {
	Record(ctx context.Context, visitorProfileID, visitedProfileID uuid.UUID) error
}

// CreateProfileInput son los datos necesarios para crear un perfil.
//
// Ya no incluye Languages ni Interests: se gestionan aparte, un ítem
// cada vez, con Service.SetLanguage / Service.SetInterest — igual que
// las fotos no se suben como parte de este struct.
type CreateProfileInput struct {
	DisplayName       string
	BirthDate         time.Time
	Gender            Gender
	CountryCode       string
	Region            *string
	RelationshipGoals []RelationshipGoal
	HasChildren       *string
	WantsChildren     *string
	Bio               *string

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

	favs   FavoriteChecker
	likes  LikeChecker
	blocks BlockChecker
	visits VisitRecorder
}

func NewService(repo Repository, store storage.Storage) *Service {
	return &Service{repo: repo, storage: store}
}

// SetInteractionDeps inyecta los colaboradores externos que hacen falta
// para completar el estado de interacción en GetFullPublicProfile.
// Se hace en un paso separado para no obligar a reordenar el wiring de
// main.go: los servicios de interacción se crean después que el de
// perfiles en el bootstrap actual.
func (s *Service) SetInteractionDeps(favs FavoriteChecker, likes LikeChecker, blocks BlockChecker, visits VisitRecorder) {
	s.favs = favs
	s.likes = likes
	s.blocks = blocks
	s.visits = visits
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

// FullProfile agrupa todo lo que useFullProfile.ts (frontend) pedía por
// separado en 6 peticiones distintas (perfil, fotos, idiomas,
// intereses, respuestas/puntuaciones de personalidad y preferencias de
// pareja) más el catálogo de intereses y el estado de interacción del
// visitante.
type FullProfile struct {
	Profile            *Profile
	Photos             []Photo
	Languages          []ProfileLanguage
	Interests          []ProfileInterest
	InterestCatalog    []InterestDefinition
	PersonalityAnswers []ProfilePersonalityAnswer
	PersonalityScores  []PersonalityTraitScore
	PartnerPreferences *PartnerPreferences

	Favorited bool
	Liked     bool
	Matched   bool
	Blocked   bool
}

// GetFullPublicProfile resuelve en una sola llamada de servicio lo que
// antes requería 6 peticiones HTTP del cliente. También carga el
// catálogo de intereses y el estado de interacción del visitante, y
// registra la visita si los colaboradores están inyectados.
func (s *Service) GetFullPublicProfile(ctx context.Context, viewerUserID, profileID uuid.UUID) (*FullProfile, error) {
	profile, err := s.repo.GetPublicByID(ctx, profileID, viewerUserID)
	if err != nil {
		return nil, err
	}

	var viewerProfile *Profile
	if s.likes != nil || s.favs != nil || s.visits != nil {
		viewerProfile, err = s.repo.GetByUserID(ctx, viewerUserID)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}

	var (
		wg                 sync.WaitGroup
		mu                 sync.Mutex
		photos             []Photo
		languages          []ProfileLanguage
		interests          []ProfileInterest
		interestCatalog    []InterestDefinition
		personalityAnswers []ProfilePersonalityAnswer
		personalityScores  []PersonalityTraitScore
		partnerPrefs       *PartnerPreferences
		firstErr           error
		favorited          bool
		liked              bool
		matched            bool
		blocked            bool
	)

	run := func(fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	run(func() (err error) { photos, err = s.repo.ListPhotos(ctx, profileID); return })
	run(func() (err error) { languages, err = s.repo.ListProfileLanguages(ctx, profileID); return })
	run(func() (err error) { interests, err = s.repo.ListProfileInterests(ctx, profileID); return })
	run(func() (err error) { interestCatalog, err = s.repo.ListInterestDefinitions(ctx); return })
	run(func() (err error) { personalityAnswers, err = s.repo.ListPersonalityAnswers(ctx, profileID); return })
	run(func() (err error) { personalityScores, err = s.repo.GetPersonalityTraitScores(ctx, profileID); return })
	run(func() (err error) { partnerPrefs, err = s.repo.GetPartnerPreferences(ctx, profileID); return })

	if s.favs != nil {
		run(func() error {
			var err error
			favorited, err = s.favs.IsFavorited(ctx, viewerUserID, profileID)
			return err
		})
	}

	if s.likes != nil && viewerProfile != nil {
		run(func() error {
			var err error
			liked, err = s.likes.IsLiked(ctx, viewerProfile.ID, profileID)
			return err
		})
		run(func() error {
			var err error
			matched, err = s.likes.HasMatch(ctx, viewerProfile.ID, profileID)
			return err
		})
	}

	if s.blocks != nil {
		run(func() error {
			var err error
			blocked, err = s.blocks.IsBlocked(ctx, viewerUserID, profile.UserID)
			return err
		})
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	if s.visits != nil && viewerProfile != nil && viewerProfile.ID != profileID {
		if err := s.visits.Record(ctx, viewerProfile.ID, profileID); err != nil {
			slog.Error("no se pudo registrar la visita", "error", err)
		}
	}

	return &FullProfile{
		Profile:            profile,
		Photos:             photos,
		Languages:          languages,
		Interests:          interests,
		InterestCatalog:    interestCatalog,
		PersonalityAnswers: personalityAnswers,
		PersonalityScores:  personalityScores,
		PartnerPreferences: partnerPrefs,
		Favorited:          favorited,
		Liked:              liked,
		Matched:            matched,
		Blocked:            blocked,
	}, nil
}

// GetMyFullProfile carga el perfil propio del usuario con todos sus
// datos agregados, pensado para la página "mi perfil". No aplica reglas
// de visibilidad pública ni registra visitas; los flags de interacción
// no tienen sentido aquí y se dejan en false.
func (s *Service) GetMyFullProfile(ctx context.Context, userID uuid.UUID) (*FullProfile, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var (
		wg                 sync.WaitGroup
		mu                 sync.Mutex
		photos             []Photo
		languages          []ProfileLanguage
		interests          []ProfileInterest
		interestCatalog    []InterestDefinition
		personalityAnswers []ProfilePersonalityAnswer
		personalityScores  []PersonalityTraitScore
		partnerPrefs       *PartnerPreferences
		firstErr           error
	)

	run := func(fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	run(func() (err error) { photos, err = s.repo.ListPhotos(ctx, profile.ID); return })
	run(func() (err error) { languages, err = s.repo.ListProfileLanguages(ctx, profile.ID); return })
	run(func() (err error) { interests, err = s.repo.ListProfileInterests(ctx, profile.ID); return })
	run(func() (err error) { interestCatalog, err = s.repo.ListInterestDefinitions(ctx); return })
	run(func() (err error) { personalityAnswers, err = s.repo.ListPersonalityAnswers(ctx, profile.ID); return })
	run(func() (err error) { personalityScores, err = s.repo.GetPersonalityTraitScores(ctx, profile.ID); return })
	run(func() (err error) { partnerPrefs, err = s.repo.GetPartnerPreferences(ctx, profile.ID); return })

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	return &FullProfile{
		Profile:            profile,
		Photos:             photos,
		Languages:          languages,
		Interests:          interests,
		InterestCatalog:    interestCatalog,
		PersonalityAnswers: personalityAnswers,
		PersonalityScores:  personalityScores,
		PartnerPreferences: partnerPrefs,
		Favorited:          false,
		Liked:              false,
		Matched:            false,
		Blocked:            false,
	}, nil
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
	if err := validateRelationshipGoals(in.RelationshipGoals); err != nil {
		return nil, err
	}
	if err := validateBio(in.Bio); err != nil {
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
		UserID:            userID,
		DisplayName:       strings.TrimSpace(in.DisplayName),
		BirthDate:         in.BirthDate,
		Gender:            in.Gender,
		CountryCode:       countryCode,
		Region:            in.Region,
		RelationshipGoals: in.RelationshipGoals,
		HasChildren:       in.HasChildren,
		WantsChildren:     in.WantsChildren,
		Bio:               in.Bio,

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
	if patch.RelationshipGoalsSet {
		if err := validateRelationshipGoals(patch.RelationshipGoals); err != nil {
			return nil, err
		}
	}
	if patch.BioSet {
		if err := validateBio(patch.Bio); err != nil {
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

// --- Idiomas del usuario -------------------------------------------------

// SetLanguage crea o actualiza el nivel de un idioma. languageCode
// inválido se traduce en un ValidationError por parte del Repository
// (violación del CHECK de profile_languages).
func (s *Service) SetLanguage(ctx context.Context, userID uuid.UUID, languageCode string, level *int) (*ProfileLanguage, error) {
	if level != nil && (*level < MinLanguageLevel || *level > MaxLanguageLevel) {
		return nil, invalidField("level", fmt.Sprintf("debe estar entre %d y %d", MinLanguageLevel, MaxLanguageLevel))
	}

	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.UpsertProfileLanguage(ctx, profile.ID, languageCode, level)
}

func (s *Service) ListMyLanguages(ctx context.Context, userID uuid.UUID) ([]ProfileLanguage, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListProfileLanguages(ctx, profile.ID)
}

func (s *Service) ListPublicLanguages(ctx context.Context, viewerUserID, profileID uuid.UUID) ([]ProfileLanguage, error) {
	if _, err := s.repo.GetPublicByID(ctx, profileID, viewerUserID); err != nil {
		return nil, err
	}
	return s.repo.ListProfileLanguages(ctx, profileID)
}

func (s *Service) DeleteLanguage(ctx context.Context, userID uuid.UUID, languageCode string) error {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.repo.DeleteProfileLanguage(ctx, profile.ID, languageCode)
}

// --- Catálogo de intereses -----------------------------------------------

func (s *Service) ListInterestCatalog(ctx context.Context) ([]InterestDefinition, error) {
	return s.repo.ListInterestDefinitions(ctx)
}

// --- Intereses del usuario -------------------------------------------------

// SetInterest crea o actualiza la respuesta del usuario a un interés
// del catálogo. Aquí se aplica la única regla de integridad que la BD
// no puede expresar porque cruza dos tablas: si el interés es de los
// que se puntúan (HasLevel=true), level es obligatorio y debe estar en
// 1-5; si no (HasLevel=false), level tiene que venir vacío — es una
// simple etiqueta presente/ausente, puntuarla no significaría nada.
func (s *Service) SetInterest(ctx context.Context, userID uuid.UUID, interestKey string, level *int) (*ProfileInterest, error) {
	def, err := s.repo.GetInterestDefinition(ctx, interestKey)
	if err != nil {
		if errors.Is(err, ErrInterestNotFound) {
			return nil, invalidField("interest_key", "no existe ese interés en el catálogo")
		}
		return nil, err
	}

	if def.HasLevel && level == nil {
		return nil, invalidField("level", "este interés requiere indicar un nivel de 1 a 5")
	}
	if !def.HasLevel && level != nil {
		return nil, invalidField("level", "este interés no admite nivel, solo puede marcarse")
	}
	if level != nil && (*level < MinInterestLevel || *level > MaxInterestLevel) {
		return nil, invalidField("level", fmt.Sprintf("debe estar entre %d y %d", MinInterestLevel, MaxInterestLevel))
	}

	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.UpsertProfileInterest(ctx, profile.ID, interestKey, level)
}

func (s *Service) ListMyInterests(ctx context.Context, userID uuid.UUID) ([]ProfileInterest, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListProfileInterests(ctx, profile.ID)
}

func (s *Service) ListPublicInterests(ctx context.Context, viewerUserID, profileID uuid.UUID) ([]ProfileInterest, error) {
	if _, err := s.repo.GetPublicByID(ctx, profileID, viewerUserID); err != nil {
		return nil, err
	}
	return s.repo.ListProfileInterests(ctx, profileID)
}

// DeleteInterest vuelve un interés a "no seleccionado". Operación idempotente.
func (s *Service) DeleteInterest(ctx context.Context, userID uuid.UUID, interestKey string) error {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.repo.DeleteProfileInterest(ctx, profile.ID, interestKey)
}

// --- Personalidad ---------------------------------------------------------

func (s *Service) ListPersonalityCatalog(ctx context.Context) ([]PersonalityStatement, error) {
	return s.repo.ListPersonalityStatements(ctx)
}

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

func (s *Service) GetPublicPartnerPreferences(ctx context.Context, viewerUserID, profileID uuid.UUID) (*PartnerPreferences, error) {
	if _, err := s.repo.GetPublicByID(ctx, profileID, viewerUserID); err != nil {
		return nil, err
	}
	return s.repo.GetPartnerPreferences(ctx, profileID)
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

// validateRelationshipGoals reemplaza a la antigua validateOptionalRelationshipGoal
// (single-select): ahora se puede marcar más de un objetivo a la vez,
// así que se valida cada uno de la lista por separado.
func validateRelationshipGoals(goals []RelationshipGoal) error {
	for _, g := range goals {
		if !allowedRelationshipGoals[g] {
			return invalidField("relationship_goals", "valor no permitido: "+string(g))
		}
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
