package search

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type resultItemResponse struct {
	ProfileID        string  `json:"profile_id"`
	DisplayName      string  `json:"display_name"`
	Age              int     `json:"age"`
	Gender           string  `json:"gender"`
	CountryCode      string  `json:"country_code"`
	Region           *string `json:"region"`
	RelationshipGoal *string `json:"relationship_goal"`
	HasPhoto         bool    `json:"has_photo"`
	CreatedAt        string  `json:"created_at"`
}

type searchResponse struct {
	Items      []resultItemResponse `json:"items"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	Total      int                  `json:"total"`
	TotalPages int                  `json:"total_pages"`
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	q := r.URL.Query()
	raw := RawQuery{
		Genders:          splitMulti(q["gender"]),
		MinAge:           q.Get("min_age"),
		MaxAge:           q.Get("max_age"),
		CountryCode:      q.Get("country"),
		Languages:        splitMulti(q["language"]),
		RelationshipGoal: q.Get("relationship_goal"),
		HasChildren:      q.Get("has_children"),
		WantsChildren:    q.Get("wants_children"),
		Interests:        splitMulti(q["interests"]),
		Page:             q.Get("page"),
		PageSize:         q.Get("page_size"),
		Sort:             q.Get("sort"),

		// --- Nuevos filtros recibidos por Query Params ---
		MinHeight:             q.Get("min_height"),
		MaxHeight:             q.Get("max_height"),
		MinWeight:             q.Get("min_weight"),
		MaxWeight:             q.Get("max_weight"),
		BodyType:              q.Get("body_type"),
		Ethnicity:             q.Get("ethnicity"),
		AppearanceRating:      q.Get("appearance_rating"),
		HairColor:             q.Get("hair_color"),
		EyeColor:              q.Get("eye_color"),
		BodyArt:               splitMulti(q["body_art"]),
		SmokingHabit:          q.Get("smoking_habit"),
		DrinkingHabit:         q.Get("drinking_habit"),
		RelocationWillingness: splitMulti(q["relocation_willingness"]),
		MaritalStatus:         q.Get("marital_status"),
		MaxChildren:           q.Get("max_children"),
		Occupation:            q.Get("occupation"),
		EmploymentStatus:      q.Get("employment_status"),
		IncomeLevel:           q.Get("income_level"),
		LivingSituation:       q.Get("living_situation"),
		Nationality:           q.Get("nationality"),
		EducationLevel:        q.Get("education_level"),
		EnglishAbility:        q.Get("english_ability"),
		Religion:              q.Get("religion"),
		ReligiousValues:       q.Get("religious_values"),
		StarSign:              q.Get("star_sign"),

		// --- NUEVOS: Hobbies y personalidad (Fase 2) -------------------
		//
		// ?hobby=cooking_baking,travelling            -> "me gusta", cualquier intensidad
		// ?hobby_travelling_min=4&hobby_gardening_max=3 -> acotan por clave
		// ?trait_openness_min=4                         -> acota un rasgo de personalidad
		//
		// La clave (hobby o rasgo) forma parte del propio nombre del
		// parámetro, así que no se puede leer con q.Get(fijo): hace
		// falta recorrer toda la query string buscando ese prefijo.
		Hobbies:                splitMulti(q["hobby"]),
		HobbyBounds:            parseKeyedBounds(q, "hobby_"),
		PersonalityTraitBounds: parseKeyedBounds(q, "trait_"),
	}

	result, err := h.svc.Search(r.Context(), userID, raw)
	if err != nil {
		var valErr *ValidationError
		if errors.As(err, &valErr) {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_param", valErr.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la búsqueda.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toSearchResponse(result))
}

func toSearchResponse(res *Result) searchResponse {
	items := make([]resultItemResponse, 0, len(res.Items))
	for _, it := range res.Items {
		var relGoal *string
		if it.RelationshipGoal != nil {
			v := string(*it.RelationshipGoal)
			relGoal = &v
		}
		items = append(items, resultItemResponse{
			ProfileID:        it.ProfileID.String(),
			DisplayName:      it.DisplayName,
			Age:              it.Age,
			Gender:           string(it.Gender),
			CountryCode:      it.CountryCode,
			Region:           it.Region,
			RelationshipGoal: relGoal,
			HasPhoto:         it.HasPhoto,
			CreatedAt:        it.CreatedAt.Format(time.RFC3339),
		})
	}

	return searchResponse{
		Items:      items,
		Page:       res.Page,
		PageSize:   res.PageSize,
		Total:      res.Total,
		TotalPages: res.TotalPages,
	}
}

func splitMulti(values []string) []string {
	var out []string
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

// parseKeyedBounds recorre toda la query string buscando parámetros con
// forma "<prefix><clave>_min" / "<prefix><clave>_max" (ej. prefix
// "hobby_" sobre "hobby_travelling_min=4") y los agrupa por clave.
//
// Se recorta el sufijo "_min"/"_max" en vez de partir por "_", porque
// la propia clave puede contener guiones bajos (hobbies como
// "cooking_baking" o "meeting_new_people"); recortar un sufijo fijo al
// final no tiene esa ambigüedad.
func parseKeyedBounds(q url.Values, prefix string) map[string]RawBounds {
	result := map[string]RawBounds{}

	for param, values := range q {
		if len(values) == 0 || values[0] == "" {
			continue
		}
		if !strings.HasPrefix(param, prefix) {
			continue
		}
		rest := strings.TrimPrefix(param, prefix)

		var key, bound string
		switch {
		case strings.HasSuffix(rest, "_min"):
			key = strings.TrimSuffix(rest, "_min")
			bound = "min"
		case strings.HasSuffix(rest, "_max"):
			key = strings.TrimSuffix(rest, "_max")
			bound = "max"
		default:
			continue
		}
		if key == "" {
			continue
		}

		b := result[key]
		if bound == "min" {
			b.Min = values[0]
		} else {
			b.Max = values[0]
		}
		result[key] = b
	}

	return result
}
