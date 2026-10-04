package search

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/httpx"
	"dating-platform/backend/internal/profiles"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// resultItemResponse es la ficha de perfil de TODOS los listados de
// búsqueda (search, new-members, online-now y popular): antes los tres
// últimos devolvían una versión reducida sin relationship_goals.
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
	// por tarjeta. PhotoThumbURL es la miniatura (480 px) de esa misma
	// foto: para las tarjetas basta y pesa mucho menos que la foto de 1600 px.
	PhotoURL      *string `json:"photo_url"`
	PhotoThumbURL *string `json:"photo_thumb_url"`
	CreatedAt     string  `json:"created_at"`

	// Relación con quien busca, resuelta en la misma consulta.
	Liked            bool `json:"liked"`
	Favorited        bool `json:"favorited"`
	ReceivedLike     bool `json:"received_like"`
	ReceivedFavorite bool `json:"received_favorite"`
}

type searchResponse struct {
	Items      []resultItemResponse `json:"items"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	Total      int                  `json:"total"`
	TotalPages int                  `json:"total_pages"`
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
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

		// Límites de nivel de intereses (has_level=true) y de rasgos de personalidad.
		InterestBounds:         parseKeyedBounds(q, "interest_"),
		PersonalityTraitBounds: parseKeyedBounds(q, "trait_"),
	}

	h.run(w, r, raw, "No se pudo completar la búsqueda.")
}

// NewMembers devuelve los perfiles más recientes.
func (h *Handler) NewMembers(w http.ResponseWriter, r *http.Request) {
	h.preset(w, r, "No se pudo obtener la lista de nuevos miembros.", func(raw *RawQuery) {
		raw.Sort = string(SortRecent)
	})
}

// OnlineNow devuelve los perfiles cuyo usuario estuvo activo en los
// últimos minutos, usando users.last_active_at.
func (h *Handler) OnlineNow(w http.ResponseWriter, r *http.Request) {
	h.preset(w, r, "No se pudo obtener la lista de usuarios en línea.", func(raw *RawQuery) {
		raw.Sort = string(SortRecent)
		raw.OnlineNow = true
	})
}

// Recommended devuelve los perfiles recientes restringidos a lo que busca
// quien mira: género buscado (obligatorio en su perfil) y rango de edad de
// sus preferencias de pareja.
func (h *Handler) Recommended(w http.ResponseWriter, r *http.Request) {
	h.preset(w, r, "No se pudo obtener la lista de recomendados.", func(raw *RawQuery) {
		raw.Sort = string(SortRecent)
		raw.UseAgePrefs = true
	})
}

// Popular devuelve los perfiles ordenados por popularidad: actividad de los
// últimos 30 días (likes, favoritos, mensajes y visitas recibidos), según la
// vista materializada profile_popularity.
func (h *Handler) Popular(w http.ResponseWriter, r *http.Request) {
	h.preset(w, r, "No se pudo obtener la lista de perfiles populares.", func(raw *RawQuery) {
		raw.Sort = string(SortPopular)
	})
}

// preset construye una búsqueda sin filtros de usuario: solo paginación más
// lo que fije `apply`. Reemplaza a los tres handlers casi idénticos de antes.
func (h *Handler) preset(w http.ResponseWriter, r *http.Request, failMsg string, apply func(*RawQuery)) {
	q := r.URL.Query()
	raw := RawQuery{
		Page:     q.Get("page"),
		PageSize: q.Get("page_size"),
	}
	apply(&raw)
	h.run(w, r, raw, failMsg)
}

func (h *Handler) run(w http.ResponseWriter, r *http.Request, raw RawQuery, failMsg string) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	result, err := h.svc.Search(r.Context(), userID, raw)
	if err != nil {
		var valErr *ValidationError
		if errors.As(err, &valErr) {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_param", valErr.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", failMsg)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toSearchResponse(result))
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

		var photoURL, photoThumbURL *string
		if it.PhotoID != nil {
			u := profiles.PublicPhotoURL(it.ProfileID, *it.PhotoID, false)
			th := profiles.PublicPhotoURL(it.ProfileID, *it.PhotoID, true)
			photoURL, photoThumbURL = &u, &th
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
			PhotoThumbURL:     photoThumbURL,
			CreatedAt:         it.CreatedAt.Format(time.RFC3339),
			Liked:             it.Liked,
			Favorited:         it.Favorited,
			ReceivedLike:      it.ReceivedLike,
			ReceivedFavorite:  it.ReceivedFavorite,
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
