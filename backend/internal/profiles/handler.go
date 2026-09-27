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
	DisplayName       string   `json:"display_name"`
	BirthDate         string   `json:"birth_date"`
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
	ar := make([]personalityAnswerResponse, 0, len(answers))
	for _, a := range answers {
		ar = append(ar, toPersonalityAnswerResponse(a))
	}
	sr := make([]personalityTraitScoreResponse, 0, len(scores))
	for _, s := range scores {
		sr = append(sr, toPersonalityTraitScoreResponse(s))
	}
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
//
// Envuelve en una sola respuesta lo que antes eran 6 peticiones
// distintas del cliente a /profiles/{id}, /photos, /languages,
// /interests, /personality y /partner-preferences. No incluye el
// catálogo de intereses (GET /catalog/interests): ese endpoint sigue
// existiendo tal cual, el cliente lo pide y cachea aparte una sola vez
// por sesión porque no depende del perfil visitado.
type fullProfileResponse struct {
	Profile            profileResponse            `json:"profile"`
	Photos             []photoResponse            `json:"photos"`
	Languages          []profileLanguageResponse  `json:"languages"`
	Interests          []profileInterestResponse  `json:"interests"`
	Personality        personalityResponse        `json:"personality"`
	PartnerPreferences partnerPreferencesResponse `json:"partner_preferences"`
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

	goals := make([]RelationshipGoal, len(req.RelationshipGoals))
	for i, g := range req.RelationshipGoals {
		goals[i] = RelationshipGoal(g)
	}

	p, err := h.svc.CreateProfile(r.Context(), userID, CreateProfileInput{
		DisplayName:       req.DisplayName,
		BirthDate:         birthDate,
		Gender:            Gender(req.Gender),
		CountryCode:       req.CountryCode,
		Region:            req.Region,
		RelationshipGoals: goals,
		HasChildren:       req.HasChildren,
		WantsChildren:     req.WantsChildren,
		Bio:               req.Bio,

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

		FutureVision:       req.FutureVision,
		Sports:             req.Sports,
		LikesPets:          req.LikesPets,
		PetsOwned:          req.PetsOwned,
		FavoriteSeason:     req.FavoriteSeason,
		IdealVacationStyle: req.IdealVacationStyle,
		VacationActivities: req.VacationActivities,
		ProfileQuote:       req.ProfileQuote,
		DreamWish:          req.DreamWish,
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

	if v, ok := raw["relationship_goals"]; ok {
		patch.RelationshipGoalsSet = true
		if !isJSONNull(v) {
			var ss []string
			if err := json.Unmarshal(v, &ss); err != nil {
				return patch, invalidField("relationship_goals", "debe ser una lista de textos o null")
			}
			goals := make([]RelationshipGoal, len(ss))
			for i, s := range ss {
				goals[i] = RelationshipGoal(s)
			}
			patch.RelationshipGoals = goals
		}
	}

	if err := parseString("region", &patch.Region, &patch.RegionSet); err != nil { return patch, err }
	if err := parseString("has_children", &patch.HasChildren, &patch.HasChildrenSet); err != nil { return patch, err }
	if err := parseString("wants_children", &patch.WantsChildren, &patch.WantsChildrenSet); err != nil { return patch, err }
	if err := parseString("bio", &patch.Bio, &patch.BioSet); err != nil { return patch, err }

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

	if err := parseSlice("future_vision", &patch.FutureVision, &patch.FutureVisionSet); err != nil { return patch, err }
	if err := parseSlice("sports", &patch.Sports, &patch.SportsSet); err != nil { return patch, err }
	if err := parseString("likes_pets", &patch.LikesPets, &patch.LikesPetsSet); err != nil { return patch, err }
	if err := parseSlice("pets_owned", &patch.PetsOwned, &patch.PetsOwnedSet); err != nil { return patch, err }
	if err := parseString("favorite_season", &patch.FavoriteSeason, &patch.FavoriteSeasonSet); err != nil { return patch, err }
	if err := parseSlice("ideal_vacation_style", &patch.IdealVacationStyle, &patch.IdealVacationStyleSet); err != nil { return patch, err }
	if err := parseSlice("vacation_activities", &patch.VacationActivities, &patch.VacationActivitiesSet); err != nil { return patch, err }
	if err := parseString("profile_quote", &patch.ProfileQuote, &patch.ProfileQuoteSet); err != nil { return patch, err }
	if err := parseString("dream_wish", &patch.DreamWish, &patch.DreamWishSet); err != nil { return patch, err }

	return patch, nil
}

func isJSONNull(raw json.RawMessage) bool {
	return string(raw) == "null"
}

// --- Fotos -------------------------------------------------------------

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

// GetPublicFull sirve GET /profiles/{profileID}/full: el equivalente
// agregado de encadenar GetPublic + ListPublicPhotos +
// ListPublicLanguages + ListPublicInterests + GetPublicPersonality +
// GetPublicPartnerPreferences, en una sola petición HTTP. Pensado para
// el cliente de Matches/Quick Match, que antes disparaba 6 peticiones
// por cada perfil mostrado.
func (h *Handler) GetPublicFull(w http.ResponseWriter, r *http.Request) {
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

	full, err := h.svc.GetFullPublicProfile(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	photos := make([]photoResponse, 0, len(full.Photos))
	for i := range full.Photos {
		photos = append(photos, toPhotoResponsePublic(&full.Photos[i], profileID))
	}

	languages := make([]profileLanguageResponse, 0, len(full.Languages))
	for _, l := range full.Languages {
		languages = append(languages, toProfileLanguageResponse(l))
	}

	interests := make([]profileInterestResponse, 0, len(full.Interests))
	for _, pi := range full.Interests {
		interests = append(interests, toProfileInterestResponse(pi))
	}

	httpx.WriteJSON(w, http.StatusOK, fullProfileResponse{
		Profile:            toProfileResponse(full.Profile),
		Photos:             photos,
		Languages:          languages,
		Interests:          interests,
		Personality:        toPersonalityResponse(full.PersonalityAnswers, full.PersonalityScores),
		PartnerPreferences: toPartnerPreferencesResponse(full.PartnerPreferences),
	})
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

// --- Handlers: idiomas del usuario -------------------------------------

func (h *Handler) ListMyLanguages(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	languages, err := h.svc.ListMyLanguages(r.Context(), userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	resp := make([]profileLanguageResponse, 0, len(languages))
	for _, l := range languages {
		resp = append(resp, toProfileLanguageResponse(l))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// SetLanguage espera el código de idioma en el path (ej. PUT
// /profiles/me/languages/{code}).
func (h *Handler) SetLanguage(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
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
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
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

	languages, err := h.svc.ListPublicLanguages(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	resp := make([]profileLanguageResponse, 0, len(languages))
	for _, l := range languages {
		resp = append(resp, toProfileLanguageResponse(l))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// --- Handlers: catálogo de intereses ------------------------------------

func (h *Handler) ListInterestCatalog(w http.ResponseWriter, r *http.Request) {
	defs, err := h.svc.ListInterestCatalog(r.Context())
	if err != nil {
		writeProfileError(w, err)
		return
	}

	resp := make([]interestDefinitionResponse, 0, len(defs))
	for _, d := range defs {
		resp = append(resp, toInterestDefinitionResponse(d))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// --- Handlers: intereses del usuario -------------------------------------

func (h *Handler) ListMyInterests(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	interests, err := h.svc.ListMyInterests(r.Context(), userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	resp := make([]profileInterestResponse, 0, len(interests))
	for _, pi := range interests {
		resp = append(resp, toProfileInterestResponse(pi))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// SetInterest espera interest_key en el path (ej. PUT
// /profiles/me/interests/{key}).
func (h *Handler) SetInterest(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
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
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
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

	interests, err := h.svc.ListPublicInterests(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	resp := make([]profileInterestResponse, 0, len(interests))
	for _, pi := range interests {
		resp = append(resp, toProfileInterestResponse(pi))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// --- Handlers: catálogo y respuestas de personalidad ----------------------

func (h *Handler) ListPersonalityCatalog(w http.ResponseWriter, r *http.Request) {
	stmts, err := h.svc.ListPersonalityCatalog(r.Context())
	if err != nil {
		writeProfileError(w, err)
		return
	}

	resp := make([]personalityStatementResponse, 0, len(stmts))
	for _, s := range stmts {
		resp = append(resp, toPersonalityStatementResponse(s))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetMyPersonality(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
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
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
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

	answers, scores, err := h.svc.GetPublicPersonality(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toPersonalityResponse(answers, scores))
}

// --- Handlers: preferencias de pareja ------------------------------------

func (h *Handler) GetMyPartnerPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
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

	patch, err := buildPartnerPreferencesPatch(raw)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	pp, err := h.svc.UpdatePartnerPreferences(r.Context(), userID, patch)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toPartnerPreferencesResponse(pp))
}

// buildPartnerPreferencesPatch traduce el JSON de la petición a un
// PartnerPreferencesPatch. Mismo patrón que buildProfilePatch: cada
// campo opcional solo se marca "Set" si su clave vino en el body,
// distinguiendo "no tocar" (clave ausente) de "borrar" (clave a null).
//
// NOTA: esta función no llegó a transmitirse en el handler.go original
// (el archivo se cortó antes de esta definición); se reconstruye aquí
// a partir del patrón de buildProfilePatch y de los campos de
// PartnerPreferencesPatch (partner_preferences.go). Verifica los
// nombres de clave JSON contra tu implementación real si difieren de
// los que ya usa partnerPreferencesResponse.
func buildPartnerPreferencesPatch(raw map[string]json.RawMessage) (PartnerPreferencesPatch, error) {
	var patch PartnerPreferencesPatch

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

	if err := parseInt("age_min", &patch.AgeMin, &patch.AgeMinSet); err != nil { return patch, err }
	if err := parseInt("age_max", &patch.AgeMax, &patch.AgeMaxSet); err != nil { return patch, err }
	if err := parseInt("height_min", &patch.HeightMin, &patch.HeightMinSet); err != nil { return patch, err }
	if err := parseInt("height_max", &patch.HeightMax, &patch.HeightMaxSet); err != nil { return patch, err }

	if err := parseSlice("desired_traits", &patch.DesiredTraits, &patch.DesiredTraitsSet); err != nil { return patch, err }

	if err := parseString("partner_may_have_children", &patch.PartnerMayHaveChildren, &patch.PartnerMayHaveChildrenSet); err != nil { return patch, err }
	if err := parseString("partner_religion_preference", &patch.PartnerReligionPreference, &patch.PartnerReligionPreferenceSet); err != nil { return patch, err }
	if err := parseString("about_partner_text", &patch.AboutPartnerText, &patch.AboutPartnerTextSet); err != nil { return patch, err }
	if err := parseString("first_meeting_preference", &patch.FirstMeetingPreference, &patch.FirstMeetingPreferenceSet); err != nil { return patch, err }
	if err := parseSlice("desired_living_place", &patch.DesiredLivingPlace, &patch.DesiredLivingPlaceSet); err != nil { return patch, err }

	if err := parseInt("importance_shared_thoughts", &patch.ImportanceSharedThoughts, &patch.ImportanceSharedThoughtsSet); err != nil { return patch, err }
	if err := parseInt("importance_shared_hobbies", &patch.ImportanceSharedHobbies, &patch.ImportanceSharedHobbiesSet); err != nil { return patch, err }
	if err := parseInt("importance_intimacy", &patch.ImportanceIntimacy, &patch.ImportanceIntimacySet); err != nil { return patch, err }
	if err := parseInt("importance_romantic_love", &patch.ImportanceRomanticLove, &patch.ImportanceRomanticLoveSet); err != nil { return patch, err }
	if err := parseInt("importance_financial_security", &patch.ImportanceFinancialSecurity, &patch.ImportanceFinancialSecuritySet); err != nil { return patch, err }
	if err := parseInt("importance_fun", &patch.ImportanceFun, &patch.ImportanceFunSet); err != nil { return patch, err }
	if err := parseInt("importance_shared_friends", &patch.ImportanceSharedFriends, &patch.ImportanceSharedFriendsSet); err != nil { return patch, err }
	if err := parseInt("importance_shared_humor", &patch.ImportanceSharedHumor, &patch.ImportanceSharedHumorSet); err != nil { return patch, err }
	if err := parseInt("importance_personal_space", &patch.ImportancePersonalSpace, &patch.ImportancePersonalSpaceSet); err != nil { return patch, err }
	if err := parseInt("importance_independence", &patch.ImportanceIndependence, &patch.ImportanceIndependenceSet); err != nil { return patch, err }

	return patch, nil
}

func (h *Handler) GetPublicPartnerPreferences(w http.ResponseWriter, r *http.Request) {
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

	pp, err := h.svc.GetPublicPartnerPreferences(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toPartnerPreferencesResponse(pp))
}

// --- Traducción de errores de dominio a respuestas HTTP -------------------
//
// ATENCIÓN — RECONSTRUCCIÓN: el archivo handler.go original se cortó al
// subirlo justo antes de llegar a estas dos funciones (se ven
// invocadas en todo el archivo, pero su definición no llegó a
// transmitirse). El cuerpo de ambas se ha reconstruido a partir de:
//   - los errores centinela definidos en errors.go
//     (ErrNotFound, ErrAlreadyExists, ErrPhotoNotFound,
//     ErrTooManyPhotos, ErrInterestNotFound, ValidationError)
//   - el patrón ya usado a mano en Handler.Get para ErrNotFound
//     (404, código "profile_not_found")
//   - el hecho de que GetPublicByID (repository.go) colapsa a
//     propósito "no existe", "cuenta inactiva" y "bloqueo" en un
//     mismo ErrNotFound, así que writePublicProfileError no debe
//     distinguir esos casos.
//
// Verifica esta implementación contra el archivo real antes de
// desplegar: en particular los códigos de error exactos
// ("profile_not_found", "invalid_field", etc.) y los status HTTP para
// ErrAlreadyExists/ErrTooManyPhotos podrían no coincidir con los que
// ya usa tu apperr/httpx si difieren de la convención de Handler.Get.

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
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Perfil no encontrado.")
		return
	}
	httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.")
}
