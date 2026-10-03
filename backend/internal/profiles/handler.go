package profiles

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/httpx"
)

const dateLayout = "2006-01-02"

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// --- DTOs: perfil -----------------------------------------------------------

type profileResponse struct {
	ID                string   `json:"id"`
	DisplayName       string   `json:"display_name"`
	Age               int      `json:"age"`
	Gender            string   `json:"gender"`
	CountryCode       string   `json:"country_code"`
	Region            *string  `json:"region"`
	RelationshipGoals []string `json:"relationship_goals"`
	HasChildren       *string  `json:"has_children"`
	WantsChildren     *string  `json:"wants_children"`
	Bio               *string  `json:"bio"`

	Height                *int     `json:"height"`
	Weight                *int     `json:"weight"`
	BodyType              *string  `json:"body_type"`
	Ethnicity             *string  `json:"ethnicity"`
	AppearanceRating      *string  `json:"appearance_rating"`
	HairColor             *string  `json:"hair_color"`
	EyeColor              *string  `json:"eye_color"`
	BodyArt               []string `json:"body_art"`
	SmokingHabit          *string  `json:"smoking_habit"`
	DrinkingHabit         *string  `json:"drinking_habit"`
	RelocationWillingness []string `json:"relocation_willingness"`
	MaritalStatus         *string  `json:"marital_status"`
	ChildrenCount         *int     `json:"children_count"`
	YoungestChildAge      *int     `json:"youngest_child_age"`
	OldestChildAge        *int     `json:"oldest_child_age"`
	Occupation            *string  `json:"occupation"`
	EmploymentStatus      *string  `json:"employment_status"`
	IncomeLevel           *string  `json:"income_level"`
	LivingSituation       *string  `json:"living_situation"`
	Nationality           *string  `json:"nationality"`
	EducationLevel        *string  `json:"education_level"`
	EnglishAbility        *string  `json:"english_ability"`
	Religion              *string  `json:"religion"`
	ReligiousValues       *string  `json:"religious_values"`
	StarSign              *string  `json:"star_sign"`

	FutureVision       []string `json:"future_vision"`
	Sports             []string `json:"sports"`
	LikesPets          *string  `json:"likes_pets"`
	PetsOwned          []string `json:"pets_owned"`
	FavoriteSeason     *string  `json:"favorite_season"`
	IdealVacationStyle []string `json:"ideal_vacation_style"`
	VacationActivities []string `json:"vacation_activities"`
	ProfileQuote       *string  `json:"profile_quote"`
	DreamWish          *string  `json:"dream_wish"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toProfileResponse(p *Profile) profileResponse {
	goals := make([]string, len(p.RelationshipGoals))
	for i, g := range p.RelationshipGoals {
		goals[i] = string(g)
	}

	return profileResponse{
		ID:                p.ID.String(),
		DisplayName:       p.DisplayName,
		Age:               p.Age(),
		Gender:            string(p.Gender),
		CountryCode:       p.CountryCode,
		Region:            p.Region,
		RelationshipGoals: goals,
		HasChildren:       p.HasChildren,
		WantsChildren:     p.WantsChildren,
		Bio:               p.Bio,

		Height:                p.Height,
		Weight:                p.Weight,
		BodyType:              p.BodyType,
		Ethnicity:             p.Ethnicity,
		AppearanceRating:      p.AppearanceRating,
		HairColor:             p.HairColor,
		EyeColor:              p.EyeColor,
		BodyArt:               p.BodyArt,
		SmokingHabit:          p.SmokingHabit,
		DrinkingHabit:         p.DrinkingHabit,
		RelocationWillingness: p.RelocationWillingness,
		MaritalStatus:         p.MaritalStatus,
		ChildrenCount:         p.ChildrenCount,
		YoungestChildAge:      p.YoungestChildAge,
		OldestChildAge:        p.OldestChildAge,
		Occupation:            p.Occupation,
		EmploymentStatus:      p.EmploymentStatus,
		IncomeLevel:           p.IncomeLevel,
		LivingSituation:       p.LivingSituation,
		Nationality:           p.Nationality,
		EducationLevel:        p.EducationLevel,
		EnglishAbility:        p.EnglishAbility,
		Religion:              p.Religion,
		ReligiousValues:       p.ReligiousValues,
		StarSign:              p.StarSign,

		FutureVision:       p.FutureVision,
		Sports:             p.Sports,
		LikesPets:          p.LikesPets,
		PetsOwned:          p.PetsOwned,
		FavoriteSeason:     p.FavoriteSeason,
		IdealVacationStyle: p.IdealVacationStyle,
		VacationActivities: p.VacationActivities,
		ProfileQuote:       p.ProfileQuote,
		DreamWish:          p.DreamWish,

		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

type photoResponse struct {
	ID string `json:"id"`
	// URL es la foto (reducida a 1600 px como máximo); ThumbURL la miniatura
	// para rejillas y listados. Si la foto es anterior a las miniaturas, el
	// servidor sirve la original también en ThumbURL.
	URL       string `json:"url"`
	ThumbURL  string `json:"thumb_url"`
	Position  int    `json:"position"`
	CreatedAt string `json:"created_at"`
}

func toPhotoResponseSelf(ph *Photo) photoResponse {
	return photoResponse{
		ID:        ph.ID.String(),
		URL:       ownPhotoURL(ph.ID, false),
		ThumbURL:  ownPhotoURL(ph.ID, true),
		Position:  ph.Position,
		CreatedAt: ph.CreatedAt.Format(time.RFC3339),
	}
}

func toPhotoResponsePublic(ph *Photo, profileID uuid.UUID) photoResponse {
	return photoResponse{
		ID:        ph.ID.String(),
		URL:       publicPhotoURL(profileID, ph.ID, false),
		ThumbURL:  publicPhotoURL(profileID, ph.ID, true),
		Position:  ph.Position,
		CreatedAt: ph.CreatedAt.Format(time.RFC3339),
	}
}

type createProfileRequest struct {
	DisplayName string `json:"display_name"`
	BirthDate   string `json:"birth_date"`
	Gender      Gender `json:"gender"`
	CountryCode string `json:"country_code"`

	ProfileDetails
}

// --- DTOs: idiomas ----------------------------------------------------

type profileLanguageResponse struct {
	LanguageCode string `json:"language_code"`
	Level        *int   `json:"level"`
	UpdatedAt    string `json:"updated_at"`
}

func toProfileLanguageResponse(pl ProfileLanguage) profileLanguageResponse {
	return profileLanguageResponse{
		LanguageCode: pl.LanguageCode,
		Level:        pl.Level,
		UpdatedAt:    pl.UpdatedAt.Format(time.RFC3339),
	}
}

type setLanguageRequest struct {
	Level *int `json:"level"`
}

// --- DTOs: intereses --------------------------------------------------

type interestDefinitionResponse struct {
	Key       string `json:"key"`
	Category  string `json:"category"`
	Label     string `json:"label"`
	HasLevel  bool   `json:"has_level"`
	SortOrder int    `json:"sort_order"`
}

func toInterestDefinitionResponse(d InterestDefinition) interestDefinitionResponse {
	return interestDefinitionResponse{Key: d.Key, Category: d.Category, Label: d.Label, HasLevel: d.HasLevel, SortOrder: d.SortOrder}
}

type profileInterestResponse struct {
	InterestKey string `json:"interest_key"`
	Level       *int   `json:"level"`
	UpdatedAt   string `json:"updated_at"`
}

func toProfileInterestResponse(pi ProfileInterest) profileInterestResponse {
	return profileInterestResponse{
		InterestKey: pi.InterestKey,
		Level:       pi.Level,
		UpdatedAt:   pi.UpdatedAt.Format(time.RFC3339),
	}
}

// setInterestRequest: level va vacío/null para los intereses concretos
// (has_level=false) y es obligatorio para los que se puntúan — el
// Service es quien conoce esa regla y la aplica, aquí solo se parsea.
type setInterestRequest struct {
	Level *int `json:"level"`
}

// --- DTOs: personalidad -------------------------------------------------

type personalityStatementResponse struct {
	Key       string `json:"key"`
	TraitKey  string `json:"trait_key"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
}

func toPersonalityStatementResponse(s PersonalityStatement) personalityStatementResponse {
	return personalityStatementResponse{Key: s.Key, TraitKey: string(s.TraitKey), Label: s.Label, SortOrder: s.SortOrder}
}

type personalityAnswerResponse struct {
	StatementKey string `json:"statement_key"`
	Score        int    `json:"score"`
	UpdatedAt    string `json:"updated_at"`
}

func toPersonalityAnswerResponse(a ProfilePersonalityAnswer) personalityAnswerResponse {
	return personalityAnswerResponse{
		StatementKey: a.StatementKey,
		Score:        a.Score,
		UpdatedAt:    a.UpdatedAt.Format(time.RFC3339),
	}
}

type personalityTraitScoreResponse struct {
	TraitKey      string  `json:"trait_key"`
	AverageScore  float64 `json:"average_score"`
	AnsweredCount int     `json:"answered_count"`
}

func toPersonalityTraitScoreResponse(s PersonalityTraitScore) personalityTraitScoreResponse {
	return personalityTraitScoreResponse{
		TraitKey:      string(s.TraitKey),
		AverageScore:  s.AverageScore,
		AnsweredCount: s.AnsweredCount,
	}
}

type personalityResponse struct {
	Answers     []personalityAnswerResponse     `json:"answers"`
	TraitScores []personalityTraitScoreResponse `json:"trait_scores"`
}

func toPersonalityResponse(answers []ProfilePersonalityAnswer, scores []PersonalityTraitScore) personalityResponse {
	ar := mapSlice(answers, toPersonalityAnswerResponse)
	sr := mapSlice(scores, toPersonalityTraitScoreResponse)
	return personalityResponse{Answers: ar, TraitScores: sr}
}

type setPersonalityAnswerRequest struct {
	Score int `json:"score"`
}

// --- DTOs: preferencias de pareja -----------------------------------------

type partnerPreferencesResponse struct {
	AgeMin    *int `json:"age_min"`
	AgeMax    *int `json:"age_max"`
	HeightMin *int `json:"height_min"`
	HeightMax *int `json:"height_max"`

	DesiredTraits []string `json:"desired_traits"`

	PartnerMayHaveChildren    *string  `json:"partner_may_have_children"`
	PartnerReligionPreference *string  `json:"partner_religion_preference"`
	AboutPartnerText          *string  `json:"about_partner_text"`
	FirstMeetingPreference    *string  `json:"first_meeting_preference"`
	DesiredLivingPlace        []string `json:"desired_living_place"`

	ImportanceSharedThoughts    *int `json:"importance_shared_thoughts"`
	ImportanceSharedHobbies     *int `json:"importance_shared_hobbies"`
	ImportanceIntimacy          *int `json:"importance_intimacy"`
	ImportanceRomanticLove      *int `json:"importance_romantic_love"`
	ImportanceFinancialSecurity *int `json:"importance_financial_security"`
	ImportanceFun               *int `json:"importance_fun"`
	ImportanceSharedFriends     *int `json:"importance_shared_friends"`
	ImportanceSharedHumor       *int `json:"importance_shared_humor"`
	ImportancePersonalSpace     *int `json:"importance_personal_space"`
	ImportanceIndependence      *int `json:"importance_independence"`

	UpdatedAt string `json:"updated_at,omitempty"`
}

func toPartnerPreferencesResponse(pp *PartnerPreferences) partnerPreferencesResponse {
	var updatedAt string
	if !pp.UpdatedAt.IsZero() {
		updatedAt = pp.UpdatedAt.Format(time.RFC3339)
	}
	return partnerPreferencesResponse{
		AgeMin: pp.AgeMin, AgeMax: pp.AgeMax, HeightMin: pp.HeightMin, HeightMax: pp.HeightMax,

		DesiredTraits: pp.DesiredTraits,

		PartnerMayHaveChildren:    pp.PartnerMayHaveChildren,
		PartnerReligionPreference: pp.PartnerReligionPreference,
		AboutPartnerText:          pp.AboutPartnerText,
		FirstMeetingPreference:    pp.FirstMeetingPreference,
		DesiredLivingPlace:        pp.DesiredLivingPlace,

		ImportanceSharedThoughts:    pp.ImportanceSharedThoughts,
		ImportanceSharedHobbies:     pp.ImportanceSharedHobbies,
		ImportanceIntimacy:          pp.ImportanceIntimacy,
		ImportanceRomanticLove:      pp.ImportanceRomanticLove,
		ImportanceFinancialSecurity: pp.ImportanceFinancialSecurity,
		ImportanceFun:               pp.ImportanceFun,
		ImportanceSharedFriends:     pp.ImportanceSharedFriends,
		ImportanceSharedHumor:       pp.ImportanceSharedHumor,
		ImportancePersonalSpace:     pp.ImportancePersonalSpace,
		ImportanceIndependence:      pp.ImportanceIndependence,

		UpdatedAt: updatedAt,
	}
}

// --- DTO: perfil completo agregado (nuevo) --------------------------------
type fullProfileResponse struct {
	Profile            profileResponse              `json:"profile"`
	Photos             []photoResponse              `json:"photos"`
	Languages          []profileLanguageResponse    `json:"languages"`
	Interests          []profileInterestResponse    `json:"interests"`
	InterestCatalog    []interestDefinitionResponse `json:"interest_catalog"`
	Personality        personalityResponse          `json:"personality"`
	PartnerPreferences partnerPreferencesResponse   `json:"partner_preferences"`

	Favorited bool `json:"favorited"`
	Liked     bool `json:"liked"`
	Matched   bool `json:"matched"`
	Blocked   bool `json:"blocked"`
}

// --- Handlers: perfil -------------------------------------------------

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	p, err := h.svc.GetMyProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Todavía no has creado tu perfil.")
			return
		}
		slog.Error("profiles: obtener perfil propio", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo obtener el perfil.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProfileResponse(p))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	var req createProfileRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	birthDate, err := time.Parse(dateLayout, req.BirthDate)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_field", "birth_date debe tener formato YYYY-MM-DD.")
		return
	}

	// Una lista de objetivos ausente se guarda como lista vacía (no NULL).
	if req.RelationshipGoals == nil {
		req.RelationshipGoals = []RelationshipGoal{}
	}

	p, err := h.svc.CreateProfile(r.Context(), userID, &Profile{
		DisplayName:    req.DisplayName,
		BirthDate:      birthDate,
		Gender:         req.Gender,
		CountryCode:    req.CountryCode,
		ProfileDetails: req.ProfileDetails,
	})
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toProfileResponse(p))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	var patch ProfilePatch
	if !decodePatchOrWrite(w, r, &patch) {
		return
	}

	p, err := h.svc.UpdateProfile(r.Context(), userID, patch)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProfileResponse(p))
}

// --- Fotos -------------------------------------------------------------

func (h *Handler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxPhotoSizeBytes+1<<20)
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "No se pudo leer el fichero enviado.")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "Falta el fichero 'photo' en el formulario.")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	photo, err := h.svc.UploadPhoto(r.Context(), userID, contentType, header.Size, file)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toPhotoResponseSelf(photo))
}

func (h *Handler) ListPhotos(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	photos, err := h.svc.ListPhotos(r.Context(), userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	resp := make([]photoResponse, 0, len(photos))
	for i := range photos {
		resp = append(resp, toPhotoResponseSelf(&photos[i]))
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) ServePhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	photoID, ok := pathUUID(w, r, "id", "foto")
	if !ok {
		return
	}

	rc, ph, err := h.svc.OpenPhoto(r.Context(), userID, photoID, wantsThumb(r))
	if err != nil {
		writeProfileError(w, err)
		return
	}
	defer rc.Close()

	serveFile(w, rc, ph.ContentType)
}

func (h *Handler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	photoID, ok := pathUUID(w, r, "id", "foto")
	if !ok {
		return
	}

	if err := h.svc.DeletePhoto(r.Context(), userID, photoID); err != nil {
		writeProfileError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetPrimaryPhoto marca una foto del usuario como principal.
func (h *Handler) SetPrimaryPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	photoID, ok := pathUUID(w, r, "id", "foto")
	if !ok {
		return
	}

	if err := h.svc.SetPrimaryPhoto(r.Context(), userID, photoID); err != nil {
		writeProfileError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// wantsThumb indica si la petición pide la miniatura (?size=thumb).
func wantsThumb(r *http.Request) bool {
	return r.URL.Query().Get("size") == "thumb"
}

// skipVisit dice si el cliente pide NO registrar la visita (?visit=0).
func skipVisit(r *http.Request) bool {
	return r.URL.Query().Get("visit") == "0"
}

// photoCacheControl: la URL de una foto lleva su UUID y el fichero no cambia,
// pero sí puede dejar de ser accesible (borrada, o bloqueo entre las dos
// personas). Por eso la caché es PRIVADA y de un día, y no un año "immutable":
// una foto ya vista no debe seguir mostrándose indefinidamente en el navegador
// de quien ya no tiene acceso. X-Content-Type-Options lo pone withSecurityHeaders.
const photoCacheControl = "private, max-age=86400"

// serveFile escribe el fichero de una foto con sus cabeceras.
func serveFile(w http.ResponseWriter, rc io.Reader, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", photoCacheControl)
	_, _ = io.Copy(w, rc)
}

func (h *Handler) GetPublic(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := requireUser(w, r)
	if !ok {
		return
	}

	profileID, ok := pathUUID(w, r, "profileID", "perfil")
	if !ok {
		return
	}

	p, err := h.svc.GetPublicProfile(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProfileResponse(p))
}

// GetMyFull sirve GET /profiles/me/full: devuelve el perfil completo del
// usuario autenticado en una sola petición. Es el endpoint que consume
// la página de "mi perfil".
func (h *Handler) GetMyFull(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	full, err := h.svc.GetMyFullProfile(r.Context(), userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toFullProfileResponse(full, toPhotoResponseSelf))
}

// GetPublicFull sirve GET /profiles/{profileID}/full: el equivalente
// agregado de encadenar GetPublic + ListPublicPhotos +
// ListPublicLanguages + ListPublicInterests + GetPublicPersonality +
// GetPublicPartnerPreferences, en una sola petición HTTP. Pensado para
// el cliente de Matches/Quick Match, que antes disparaba 6 peticiones
// por cada perfil mostrado. Ahora también devuelve el catálogo de
// intereses y los flags de interacción del visitante.
func (h *Handler) GetPublicFull(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := requireUser(w, r)
	if !ok {
		return
	}

	profileID, ok := pathUUID(w, r, "profileID", "perfil")
	if !ok {
		return
	}

	full, err := h.svc.GetFullPublicProfile(r.Context(), viewerID, profileID, FullPublicOptions{SkipVisit: skipVisit(r)})
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	toPhoto := func(ph *Photo) photoResponse { return toPhotoResponsePublic(ph, profileID) }
	httpx.WriteJSON(w, http.StatusOK, toFullProfileResponse(full, toPhoto))
}

// toFullProfileResponse arma la respuesta agregada de /full. toPhoto decide la
// URL de las fotos (propias o públicas); los flags de interacción vienen de
// full (en el perfil propio son siempre false).
func toFullProfileResponse(full *FullProfile, toPhoto func(*Photo) photoResponse) fullProfileResponse {
	photos := make([]photoResponse, 0, len(full.Photos))
	for i := range full.Photos {
		photos = append(photos, toPhoto(&full.Photos[i]))
	}

	return fullProfileResponse{
		Profile:            toProfileResponse(full.Profile),
		Photos:             photos,
		Languages:          mapSlice(full.Languages, toProfileLanguageResponse),
		Interests:          mapSlice(full.Interests, toProfileInterestResponse),
		InterestCatalog:    mapSlice(full.InterestCatalog, toInterestDefinitionResponse),
		Personality:        toPersonalityResponse(full.PersonalityAnswers, full.PersonalityScores),
		PartnerPreferences: toPartnerPreferencesResponse(full.PartnerPreferences),
		Favorited:          full.Favorited,
		Liked:              full.Liked,
		Matched:            full.Matched,
		Blocked:            false, // un bloqueo hace el perfil invisible (404): nunca se llega aquí bloqueado
	}
}

func (h *Handler) ListPublicPhotos(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := requireUser(w, r)
	if !ok {
		return
	}

	profileID, ok := pathUUID(w, r, "profileID", "perfil")
	if !ok {
		return
	}

	photos, err := h.svc.ListPublicPhotos(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	resp := make([]photoResponse, 0, len(photos))
	for i := range photos {
		resp = append(resp, toPhotoResponsePublic(&photos[i], profileID))
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) ServePublicPhoto(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := requireUser(w, r)
	if !ok {
		return
	}

	profileID, ok := pathUUID(w, r, "profileID", "perfil")
	if !ok {
		return
	}
	photoID, ok := pathUUID(w, r, "photoID", "foto")
	if !ok {
		return
	}

	rc, ph, err := h.svc.OpenPublicPhoto(r.Context(), viewerID, profileID, photoID, wantsThumb(r))
	if err != nil {
		writePublicProfileError(w, err)
		return
	}
	defer rc.Close()

	serveFile(w, rc, ph.ContentType)
}

// --- Handlers: idiomas del usuario -------------------------------------

func (h *Handler) ListMyLanguages(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	languages, err := h.svc.ListMyLanguages(r.Context(), userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	resp := mapSlice(languages, toProfileLanguageResponse)
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// SetLanguage espera el código de idioma en el path (ej. PUT
// /profiles/me/languages/{code}).
func (h *Handler) SetLanguage(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	code := r.PathValue("code")

	var req setLanguageRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	pl, err := h.svc.SetLanguage(r.Context(), userID, code, req.Level)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProfileLanguageResponse(*pl))
}

func (h *Handler) DeleteLanguage(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	code := r.PathValue("code")

	if err := h.svc.DeleteLanguage(r.Context(), userID, code); err != nil {
		writeProfileError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListPublicLanguages(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := requireUser(w, r)
	if !ok {
		return
	}

	profileID, ok := pathUUID(w, r, "profileID", "perfil")
	if !ok {
		return
	}

	languages, err := h.svc.ListPublicLanguages(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	resp := mapSlice(languages, toProfileLanguageResponse)
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// --- Handlers: catálogo de intereses ------------------------------------

func (h *Handler) ListInterestCatalog(w http.ResponseWriter, r *http.Request) {
	defs, err := h.svc.ListInterestCatalog(r.Context())
	if err != nil {
		writeProfileError(w, err)
		return
	}

	resp := mapSlice(defs, toInterestDefinitionResponse)
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// --- Handlers: intereses del usuario -------------------------------------

func (h *Handler) ListMyInterests(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	interests, err := h.svc.ListMyInterests(r.Context(), userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	resp := mapSlice(interests, toProfileInterestResponse)
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// SetInterest espera interest_key en el path (ej. PUT
// /profiles/me/interests/{key}).
func (h *Handler) SetInterest(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	interestKey := r.PathValue("key")

	var req setInterestRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	pi, err := h.svc.SetInterest(r.Context(), userID, interestKey, req.Level)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProfileInterestResponse(*pi))
}

// DeleteInterest vuelve un interés a "no seleccionado" (idempotente).
func (h *Handler) DeleteInterest(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	interestKey := r.PathValue("key")

	if err := h.svc.DeleteInterest(r.Context(), userID, interestKey); err != nil {
		writeProfileError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListPublicInterests(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := requireUser(w, r)
	if !ok {
		return
	}

	profileID, ok := pathUUID(w, r, "profileID", "perfil")
	if !ok {
		return
	}

	interests, err := h.svc.ListPublicInterests(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	resp := mapSlice(interests, toProfileInterestResponse)
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// --- Handlers: catálogo y respuestas de personalidad ----------------------

func (h *Handler) ListPersonalityCatalog(w http.ResponseWriter, r *http.Request) {
	stmts, err := h.svc.ListPersonalityCatalog(r.Context())
	if err != nil {
		writeProfileError(w, err)
		return
	}

	resp := mapSlice(stmts, toPersonalityStatementResponse)
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetMyPersonality(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	answers, scores, err := h.svc.GetMyPersonality(r.Context(), userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toPersonalityResponse(answers, scores))
}

func (h *Handler) SetPersonalityAnswer(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	statementKey := r.PathValue("key")

	var req setPersonalityAnswerRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	a, err := h.svc.SetPersonalityAnswer(r.Context(), userID, statementKey, req.Score)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toPersonalityAnswerResponse(*a))
}

func (h *Handler) GetPublicPersonality(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := requireUser(w, r)
	if !ok {
		return
	}

	profileID, ok := pathUUID(w, r, "profileID", "perfil")
	if !ok {
		return
	}

	answers, scores, err := h.svc.GetPublicPersonality(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toPersonalityResponse(answers, scores))
}

// --- Handlers: preferencias de pareja ------------------------------------

func (h *Handler) GetMyPartnerPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	pp, err := h.svc.GetMyPartnerPreferences(r.Context(), userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toPartnerPreferencesResponse(pp))
}

func (h *Handler) UpdatePartnerPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	var patch PartnerPreferencesPatch
	if !decodePatchOrWrite(w, r, &patch) {
		return
	}

	pp, err := h.svc.UpdatePartnerPreferences(r.Context(), userID, patch)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toPartnerPreferencesResponse(pp))
}

func (h *Handler) GetPublicPartnerPreferences(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := requireUser(w, r)
	if !ok {
		return
	}

	profileID, ok := pathUUID(w, r, "profileID", "perfil")
	if !ok {
		return
	}

	pp, err := h.svc.GetPublicPartnerPreferences(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toPartnerPreferencesResponse(pp))
}

// --- Traducción de errores de dominio a respuestas HTTP -------------------

// writeProfileError traduce los errores de dominio de profiles al
// código y mensaje HTTP correspondientes, para los endpoints
// autenticados sobre el propio perfil (crear, actualizar, fotos,
// idiomas, intereses, personalidad, preferencias de pareja).
func writeProfileError(w http.ResponseWriter, err error) {
	var valErr *ValidationError
	switch {
	case errors.As(err, &valErr):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_field", valErr.Error())
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Perfil no encontrado.")
	case errors.Is(err, ErrAlreadyExists):
		httpx.WriteError(w, http.StatusConflict, "profile_already_exists", "Ya tienes un perfil creado.")
	case errors.Is(err, ErrPhotoNotFound):
		httpx.WriteError(w, http.StatusNotFound, "photo_not_found", "Foto no encontrada.")
	case errors.Is(err, ErrTooManyPhotos):
		httpx.WriteError(w, http.StatusBadRequest, "too_many_photos", "Se alcanzó el número máximo de fotos.")
	case errors.Is(err, ErrInterestNotFound):
		httpx.WriteError(w, http.StatusBadRequest, "interest_not_found", "Interés no encontrado en el catálogo.")
	default:
		slog.Error("profiles: error no controlado", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.")
	}
}

// writePublicProfileError es la variante para endpoints públicos
// (GetPublic, GetPublicFull, ListPublicPhotos, ListPublicLanguages,
// ListPublicInterests, GetPublicPersonality,
// GetPublicPartnerPreferences). Devuelve siempre un 404 genérico ante
// ErrNotFound, sin distinguir "no existe" de "bloqueado" (ver nota de
// reconstrucción arriba).
func writePublicProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Perfil no encontrado.")
		return
	case errors.Is(err, ErrPhotoNotFound):
		httpx.WriteError(w, http.StatusNotFound, "photo_not_found", "Foto no encontrada.")
		return
	}
	slog.Error("profiles: error no controlado (público)", "error", err)
	httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.")
}
