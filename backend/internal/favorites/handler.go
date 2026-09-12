package favorites

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

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

type favoriteItemResponse struct {
	ProfileID        string  `json:"profile_id"`
	DisplayName      string  `json:"display_name"`
	Age              int     `json:"age"`
	Gender           string  `json:"gender"`
	CountryCode      string  `json:"country_code"`
	Region           *string `json:"region"`
	RelationshipGoal *string `json:"relationship_goal"`
	HasPhoto         bool    `json:"has_photo"`
	FavoritedAt      string  `json:"favorited_at"`
}

type listResponse struct {
	Items      []favoriteItemResponse `json:"items"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	Total      int                    `json:"total"`
	TotalPages int                    `json:"total_pages"`
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	userID, profileID, ok := h.authAndProfileID(w, r)
	if !ok {
		return
	}

	if err := h.svc.Add(r.Context(), userID, profileID); err != nil {
		writeFavoritesError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, profileID, ok := h.authAndProfileID(w, r)
	if !ok {
		return
	}

	if err := h.svc.Remove(r.Context(), userID, profileID); err != nil {
		writeFavoritesError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	userID, profileID, ok := h.authAndProfileID(w, r)
	if !ok {
		return
	}

	favorited, err := h.svc.IsFavorited(r.Context(), userID, profileID)
	if err != nil {
		writeFavoritesError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"favorited": favorited})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	page := 1
	if v := r.URL.Query().Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_param", "page debe ser un entero >= 1.")
			return
		}
		page = n
	}

	pageSize := DefaultPageSize
	if v := r.URL.Query().Get("page_size"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_param", "page_size debe ser un entero >= 1.")
			return
		}
		pageSize = n
	}

	result, err := h.svc.List(r.Context(), userID, page, pageSize)
	if err != nil {
		writeFavoritesError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toListResponse(result))
}

// authAndProfileID hace las dos comprobaciones que repiten Add/Remove/Status:
// que haya sesión y que {profileID} en la ruta sea un UUID válido.
func (h *Handler) authAndProfileID(w http.ResponseWriter, r *http.Request) (userID, profileID uuid.UUID, ok bool) {
	userID, ok = auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return uuid.Nil, uuid.Nil, false
	}

	profileID, err := uuid.Parse(r.PathValue("profileID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de perfil inválido.")
		return uuid.Nil, uuid.Nil, false
	}

	return userID, profileID, true
}

func toListResponse(res *ListResult) listResponse {
	items := make([]favoriteItemResponse, 0, len(res.Items))
	for _, it := range res.Items {
		var relGoal *string
		if it.RelationshipGoal != nil {
			v := string(*it.RelationshipGoal)
			relGoal = &v
		}
		items = append(items, favoriteItemResponse{
			ProfileID:        it.ProfileID.String(),
			DisplayName:      it.DisplayName,
			Age:              it.Age,
			Gender:           string(it.Gender),
			CountryCode:      it.CountryCode,
			Region:           it.Region,
			RelationshipGoal: relGoal,
			HasPhoto:         it.HasPhoto,
			FavoritedAt:      it.FavoritedAt.Format(time.RFC3339),
		})
	}

	return listResponse{
		Items:      items,
		Page:       res.Page,
		PageSize:   res.PageSize,
		Total:      res.Total,
		TotalPages: res.TotalPages,
	}
}

func writeFavoritesError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrCannotFavoriteSelf):
		httpx.WriteError(w, http.StatusBadRequest, "cannot_favorite_self", "No puedes añadirte a ti mismo a favoritos.")
	case errors.Is(err, profiles.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Perfil no encontrado.")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.")
	}
}
