package profiles

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/httpx"
)

const dateLayout = "2006-01-02"

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// --- DTOs -----------------------------------------------------------------

type profileResponse struct {
	ID               string   `json:"id"`
	DisplayName      string   `json:"display_name"`
	Age              int      `json:"age"`
	Gender           string   `json:"gender"`
	CountryCode      string   `json:"country_code"`
	Region           *string  `json:"region"`
	Languages        []string `json:"languages"`
	RelationshipGoal *string  `json:"relationship_goal"`
	HasChildren      *string  `json:"has_children"`
	WantsChildren    *string  `json:"wants_children"`
	Bio              *string  `json:"bio"`
	Interests        []string `json:"interests"`

	// --- Nuevos campos ---
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

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toProfileResponse(p *Profile) profileResponse {
	var relGoal *string
	if p.RelationshipGoal != nil {
		v := string(*p.RelationshipGoal)
		relGoal = &v
	}
	return profileResponse{
		ID:               p.ID.String(),
		DisplayName:      p.DisplayName,
		Age:              p.Age(),
		Gender:           string(p.Gender),
		CountryCode:      p.CountryCode,
		Region:           p.Region,
		Languages:        p.Languages,
		RelationshipGoal: relGoal,
		HasChildren:      p.HasChildren,
		WantsChildren:    p.WantsChildren,
		Bio:              p.Bio,
		Interests:        p.Interests,

		// Mapeo de campos nuevos
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

		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

type photoResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Position  int    `json:"position"`
	CreatedAt string `json:"created_at"`
}

func toPhotoResponseSelf(ph *Photo) photoResponse {
	return photoResponse{
		ID:        ph.ID.String(),
		URL:       fmt.Sprintf("/api/v1/profiles/me/photos/%s/file", ph.ID.String()),
		Position:  ph.Position,
		CreatedAt: ph.CreatedAt.Format(time.RFC3339),
	}
}

func toPhotoResponsePublic(ph *Photo, profileID uuid.UUID) photoResponse {
	return photoResponse{
		ID:        ph.ID.String(),
		URL:       fmt.Sprintf("/api/v1/profiles/%s/photos/%s/file", profileID.String(), ph.ID.String()),
		Position:  ph.Position,
		CreatedAt: ph.CreatedAt.Format(time.RFC3339),
	}
}

type createProfileRequest struct {
	DisplayName      string   `json:"display_name"`
	BirthDate        string   `json:"birth_date"`
	Gender           string   `json:"gender"`
	CountryCode      string   `json:"country_code"`
	Region           *string  `json:"region"`
	Languages        []string `json:"languages"`
	RelationshipGoal *string  `json:"relationship_goal"`
	HasChildren      *string  `json:"has_children"`
	WantsChildren    *string  `json:"wants_children"`
	Bio              *string  `json:"bio"`
	Interests        []string `json:"interests"`

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
}

// --- Handlers: perfil -------------------------------------------------

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	p, err := h.svc.GetMyProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Todavía no has creado tu perfil.")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo obtener el perfil.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProfileResponse(p))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
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

	var relGoal *RelationshipGoal
	if req.RelationshipGoal != nil {
		g := RelationshipGoal(*req.RelationshipGoal)
		relGoal = &g
	}

	p, err := h.svc.CreateProfile(r.Context(), userID, CreateProfileInput{
		DisplayName:      req.DisplayName,
		BirthDate:        birthDate,
		Gender:           Gender(req.Gender),
		CountryCode:      req.CountryCode,
		Region:           req.Region,
		Languages:        req.Languages,
		RelationshipGoal: relGoal,
		HasChildren:      req.HasChildren,
		WantsChildren:    req.WantsChildren,
		Bio:              req.Bio,
		Interests:        req.Interests,

		Height:                req.Height,
		Weight:                req.Weight,
		BodyType:              req.BodyType,
		Ethnicity:             req.Ethnicity,
		AppearanceRating:      req.AppearanceRating,
		HairColor:             req.HairColor,
		EyeColor:              req.EyeColor,
		BodyArt:               req.BodyArt,
		SmokingHabit:          req.SmokingHabit,
		DrinkingHabit:         req.DrinkingHabit,
		RelocationWillingness: req.RelocationWillingness,
		MaritalStatus:         req.MaritalStatus,
		ChildrenCount:         req.ChildrenCount,
		YoungestChildAge:      req.YoungestChildAge,
		OldestChildAge:        req.OldestChildAge,
		Occupation:            req.Occupation,
		EmploymentStatus:      req.EmploymentStatus,
		IncomeLevel:           req.IncomeLevel,
		LivingSituation:       req.LivingSituation,
		Nationality:           req.Nationality,
		EducationLevel:        req.EducationLevel,
		EnglishAbility:        req.EnglishAbility,
		Religion:              req.Religion,
		ReligiousValues:       req.ReligiousValues,
		StarSign:              req.StarSign,
	})
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toProfileResponse(p))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	var raw map[string]json.RawMessage
	if err := httpx.DecodeJSON(r, &raw); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	patch, err := buildProfilePatch(raw)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	p, err := h.svc.UpdateProfile(r.Context(), userID, patch)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProfileResponse(p))
}

// buildProfilePatch traduce el JSON de la petición a un ProfilePatch
func buildProfilePatch(raw map[string]json.RawMessage) (ProfilePatch, error) {
	var patch ProfilePatch

	parseString := func(key string, target **string, flag *bool) error {
		if v, ok := raw[key]; ok {
			*flag = true
			if !isJSONNull(v) {
				var s string
				if err := json.Unmarshal(v, &s); err != nil {
					return invalidField(key, "debe ser texto o null")
				}
				*target = &s
			}
		}
		return nil
	}

	parseInt := func(key string, target **int, flag *bool) error {
		if v, ok := raw[key]; ok {
			*flag = true
			if !isJSONNull(v) {
				var i int
				if err := json.Unmarshal(v, &i); err != nil {
					return invalidField(key, "debe ser un entero o null")
				}
				*target = &i
			}
		}
		return nil
	}

	parseSlice := func(key string, target *[]string, flag *bool) error {
		if v, ok := raw[key]; ok {
			*flag = true
			if !isJSONNull(v) {
				var sl []string
				if err := json.Unmarshal(v, &sl); err != nil {
					return invalidField(key, "debe ser una lista de textos o null")
				}
				*target = sl
			}
		}
		return nil
	}

	if v, ok := raw["display_name"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return patch, invalidField("display_name", "debe ser texto")
		}
		patch.DisplayName = &s
	}

	if v, ok := raw["birth_date"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return patch, invalidField("birth_date", "debe ser texto")
		}
		t, err := time.Parse(dateLayout, s)
		if err != nil {
			return patch, invalidField("birth_date", "formato esperado YYYY-MM-DD")
		}
		patch.BirthDate = &t
	}

	if v, ok := raw["gender"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return patch, invalidField("gender", "debe ser texto")
		}
		g := Gender(s)
		patch.Gender = &g
	}

	if v, ok := raw["country_code"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return patch, invalidField("country_code", "debe ser texto")
		}
		patch.CountryCode = &s
	}

	if v, ok := raw["relationship_goal"]; ok {
		patch.RelationshipGoalSet = true
		if !isJSONNull(v) {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return patch, invalidField("relationship_goal", "debe ser texto o null")
			}
			g := RelationshipGoal(s)
			patch.RelationshipGoal = &g
		}
	}

	// Parsing de los campos opcionales mediante helpers
	if err := parseString("region", &patch.Region, &patch.RegionSet); err != nil { return patch, err }
	if err := parseSlice("languages", &patch.Languages, &patch.LanguagesSet); err != nil { return patch, err }
	if err := parseString("has_children", &patch.HasChildren, &patch.HasChildrenSet); err != nil { return patch, err }
	if err := parseString("wants_children", &patch.WantsChildren, &patch.WantsChildrenSet); err != nil { return patch, err }
	if err := parseString("bio", &patch.Bio, &patch.BioSet); err != nil { return patch, err }
	if err := parseSlice("interests", &patch.Interests, &patch.InterestsSet); err != nil { return patch, err }

	// Nuevos campos
	if err := parseInt("height", &patch.Height, &patch.HeightSet); err != nil { return patch, err }
	if err := parseInt("weight", &patch.Weight, &patch.WeightSet); err != nil { return patch, err }
	if err := parseString("body_type", &patch.BodyType, &patch.BodyTypeSet); err != nil { return patch, err }
	if err := parseString("ethnicity", &patch.Ethnicity, &patch.EthnicitySet); err != nil { return patch, err }
	if err := parseString("appearance_rating", &patch.AppearanceRating, &patch.AppearanceRatingSet); err != nil { return patch, err }
	if err := parseString("hair_color", &patch.HairColor, &patch.HairColorSet); err != nil { return patch, err }
	if err := parseString("eye_color", &patch.EyeColor, &patch.EyeColorSet); err != nil { return patch, err }
	if err := parseSlice("body_art", &patch.BodyArt, &patch.BodyArtSet); err != nil { return patch, err }

	if err := parseString("smoking_habit", &patch.SmokingHabit, &patch.SmokingHabitSet); err != nil { return patch, err }
	if err := parseString("drinking_habit", &patch.DrinkingHabit, &patch.DrinkingHabitSet); err != nil { return patch, err }
	if err := parseSlice("relocation_willingness", &patch.RelocationWillingness, &patch.RelocationWillingnessSet); err != nil { return patch, err }
	if err := parseString("marital_status", &patch.MaritalStatus, &patch.MaritalStatusSet); err != nil { return patch, err }
	if err := parseInt("children_count", &patch.ChildrenCount, &patch.ChildrenCountSet); err != nil { return patch, err }
	if err := parseInt("youngest_child_age", &patch.YoungestChildAge, &patch.YoungestChildAgeSet); err != nil { return patch, err }
	if err := parseInt("oldest_child_age", &patch.OldestChildAge, &patch.OldestChildAgeSet); err != nil { return patch, err }
	if err := parseString("occupation", &patch.Occupation, &patch.OccupationSet); err != nil { return patch, err }
	if err := parseString("employment_status", &patch.EmploymentStatus, &patch.EmploymentStatusSet); err != nil { return patch, err }
	if err := parseString("income_level", &patch.IncomeLevel, &patch.IncomeLevelSet); err != nil { return patch, err }
	if err := parseString("living_situation", &patch.LivingSituation, &patch.LivingSituationSet); err != nil { return patch, err }

	if err := parseString("nationality", &patch.Nationality, &patch.NationalitySet); err != nil { return patch, err }
	if err := parseString("education_level", &patch.EducationLevel, &patch.EducationLevelSet); err != nil { return patch, err }
	if err := parseString("english_ability", &patch.EnglishAbility, &patch.EnglishAbilitySet); err != nil { return patch, err }
	if err := parseString("religion", &patch.Religion, &patch.ReligionSet); err != nil { return patch, err }
	if err := parseString("religious_values", &patch.ReligiousValues, &patch.ReligiousValuesSet); err != nil { return patch, err }
	if err := parseString("star_sign", &patch.StarSign, &patch.StarSignSet); err != nil { return patch, err }

	return patch, nil
}

func isJSONNull(raw json.RawMessage) bool {
	return string(raw) == "null"
}

// --- Fotos y perfiles públicos se mantienen idénticos ---

func (h *Handler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
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
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
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
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	photoID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de foto inválido.")
		return
	}

	rc, ph, err := h.svc.OpenPhoto(r.Context(), userID, photoID)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", ph.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	_, _ = io.Copy(w, rc)
}

func (h *Handler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	photoID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de foto inválido.")
		return
	}

	if err := h.svc.DeletePhoto(r.Context(), userID, photoID); err != nil {
		writeProfileError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetPublic(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	profileID, err := uuid.Parse(r.PathValue("profileID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de perfil inválido.")
		return
	}

	p, err := h.svc.GetPublicProfile(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProfileResponse(p))
}

func (h *Handler) ListPublicPhotos(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	profileID, err := uuid.Parse(r.PathValue("profileID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de perfil inválido.")
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
	viewerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	profileID, err := uuid.Parse(r.PathValue("profileID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de perfil inválido.")
		return
	}
	photoID, err := uuid.Parse(r.PathValue("photoID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de foto inválido.")
		return
	}

	rc, ph, err := h.svc.OpenPublicPhoto(r.Context(), viewerID, profileID, photoID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", ph.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	_, _ = io.Copy(w, rc)
}

// --- Errores --------------------------------------------------------------

func writePublicProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Perfil no encontrado.")
	case errors.Is(err, ErrPhotoNotFound):
		httpx.WriteError(w, http.StatusNotFound, "photo_not_found", "Foto no encontrada.")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.")
	}
}

func writeProfileError(w http.ResponseWriter, err error) {
	var valErr *ValidationError
	switch {
	case errors.As(err, &valErr):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_field", valErr.Error())
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Todavía no has creado tu perfil.")
	case errors.Is(err, ErrAlreadyExists):
		httpx.WriteError(w, http.StatusConflict, "profile_already_exists", "Ya tienes un perfil creado.")
	case errors.Is(err, ErrPhotoNotFound):
		httpx.WriteError(w, http.StatusNotFound, "photo_not_found", "Foto no encontrada.")
	case errors.Is(err, ErrTooManyPhotos):
		httpx.WriteError(w, http.StatusConflict, "too_many_photos", fmt.Sprintf("Máximo %d fotos por perfil.", MaxPhotosPerProfile))
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.")
	}
}