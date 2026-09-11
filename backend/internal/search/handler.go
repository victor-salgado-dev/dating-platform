package search

import (
	"errors"
	"net/http"
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

// splitMulti aplana valores de query string repetidos (?gender=a&gender=b)
// y/o separados por comas (?gender=a,b) en una sola lista limpia.
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
