package search

import (
	"errors"
	"fmt"
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
	ProfileID         string   `json:"profile_id"`
	DisplayName       string   `json:"display_name"`
	Age               int      `json:"age"`
	Gender            string   `json:"gender"`
	CountryCode       string   `json:"country_code"`
	Region            *string  `json:"region"`
	RelationshipGoal  *string  `json:"relationship_goal,omitempty"`
	RelationshipGoals []string `json:"relationship_goals"`
	HasPhoto          bool     `json:"has_photo"`
	// PhotoURL es la URL ya armada de la foto principal del perfil, o
	// null si no tiene ninguna. Se resuelve en la misma consulta que
	// trae la lista, para que el grid del cliente no haga una petición
	// por tarjeta.
	PhotoURL  *string `json:"photo_url"`
	CreatedAt string  `json:"created_at"`
}

type searchResponse struct {
	Items      []resultItemResponse `json:"items"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	Total      int                  `json:"total"`
	TotalPages int                  `json:"total_pages"`
}

// newMemberItemResponse es la versión reducida para el endpoint
// "new members": solo los datos que necesita una tarjeta simple.
type newMemberItemResponse struct {
	ProfileID   string  `json:"profile_id"`
	DisplayName string  `json:"display_name"`
	Age         int     `json:"age"`
	Gender      string  `json:"gender"`
	CountryCode string  `json:"country_code"`
	Region      *string `json:"region"`
	PhotoURL    *string `json:"photo_url"`
	CreatedAt   string  `json:"created_at"`
}

type newMembersResponse struct {
	Items      []newMemberItemResponse `json:"items"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	Total      int                     `json:"total"`
	TotalPages int                     `json:"total_pages"`
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	q := r.URL.Query()
	raw := RawQuery{
		Genders:           splitMulti(q["gender"]),
		MinAge:            q.Get("min_age"),
		MaxAge:            q.Get("max_age"),
		CountryCode:       q.Get("country"),
		Languages:         splitMulti(q["language"]),
		RelationshipGoals: append(splitMulti(q["relationship_goals"]), splitMulti(q["relationship_goal"])...),
		HasChildren:       q.Get("has_children"),
		WantsChildren:     q.Get("wants_children"),
		Interests:         splitMulti(q["interests"]),
		Page:              q.Get("page"),
		PageSize:          q.Get("page_size"),
		Sort:              q.Get("sort"),

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

		// --- Estilo de vida adicional ---
		FutureVision:       splitMulti(q["future_vision"]),
		Sports:             splitMulti(q["sports"]),
		LikesPets:          q.Get("likes_pets"),
		PetsOwned:          splitMulti(q["pets_owned"]),
		FavoriteSeason:     q.Get("favorite_season"),
		IdealVacationStyle: splitMulti(q["ideal_vacation_style"]),
		VacationActivities: splitMulti(q["vacation_activities"]),

		// --- Nivel de intereses (has_level=true) ---
		InterestBounds: parseKeyedBounds(q, "interest_"),

		// --- Personalidad (Fase 2) ---
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

// NewMembers devuelve los perfiles más recientes, usando la misma capa
// de búsqueda/configuración de visibilidad que Search, pero con una
// respuesta más compacta.
func (h *Handler) NewMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	q := r.URL.Query()
	raw := RawQuery{
		Page:     q.Get("page"),
		PageSize: q.Get("page_size"),
		Sort:     string(SortRecent),
	}

	result, err := h.svc.Search(r.Context(), userID, raw)
	if err != nil {
		var valErr *ValidationError
		if errors.As(err, &valErr) {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_param", valErr.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo obtener la lista de nuevos miembros.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toNewMembersResponse(result))
}

// OnlineNow devuelve los perfiles cuyo usuario estuvo activo en los
// últimos minutos, usando last_active_at.
func (h *Handler) OnlineNow(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	q := r.URL.Query()
	raw := RawQuery{
		Page:      q.Get("page"),
		PageSize:  q.Get("page_size"),
		Sort:      string(SortRecent),
		OnlineNow: true,
	}

	result, err := h.svc.Search(r.Context(), userID, raw)
	if err != nil {
		var valErr *ValidationError
		if errors.As(err, &valErr) {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_param", valErr.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo obtener la lista de usuarios en línea.")
		return
	}

	// La respuesta reducida de new-members cumple exactamente el contrato
	// simple que el frontend necesita para la pestaña Online.
	httpx.WriteJSON(w, http.StatusOK, toNewMembersResponse(result))
}

func toSearchResponse(res *Result) searchResponse {
	items := make([]resultItemResponse, 0, len(res.Items))
	for _, it := range res.Items {
		goals := make([]string, len(it.RelationshipGoals))
		for i, g := range it.RelationshipGoals {
			goals[i] = string(g)
		}

		var firstGoal *string
		if len(goals) > 0 {
			firstGoal = &goals[0]
		}

		var photoURL *string
		if it.PhotoID != nil {
			u := fmt.Sprintf("/api/v1/profiles/%s/photos/%s/file",
				it.ProfileID.String(), it.PhotoID.String())
			photoURL = &u
		}

		items = append(items, resultItemResponse{
			ProfileID:         it.ProfileID.String(),
			DisplayName:       it.DisplayName,
			Age:               it.Age,
			Gender:            string(it.Gender),
			CountryCode:       it.CountryCode,
			Region:            it.Region,
			RelationshipGoal:  firstGoal,
			RelationshipGoals: goals,
			HasPhoto:          it.HasPhoto,
			PhotoURL:          photoURL,
			CreatedAt:         it.CreatedAt.Format(time.RFC3339),
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

func toNewMembersResponse(res *Result) newMembersResponse {
	items := make([]newMemberItemResponse, 0, len(res.Items))
	for _, it := range res.Items {
		var photoURL *string
		if it.PhotoID != nil {
			u := fmt.Sprintf("/api/v1/profiles/%s/photos/%s/file",
				it.ProfileID.String(), it.PhotoID.String())
			photoURL = &u
		}

		items = append(items, newMemberItemResponse{
			ProfileID:   it.ProfileID.String(),
			DisplayName: it.DisplayName,
			Age:         it.Age,
			Gender:      string(it.Gender),
			CountryCode: it.CountryCode,
			Region:      it.Region,
			PhotoURL:    photoURL,
			CreatedAt:   it.CreatedAt.Format(time.RFC3339),
		})
	}

	return newMembersResponse{
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
