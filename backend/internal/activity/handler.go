package activity

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/httpx"
	"dating-platform/backend/internal/profiles"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type itemResponse struct {
	EventType   string  `json:"event_type"`
	ProfileID   string  `json:"profile_id"`
	DisplayName string  `json:"display_name"`
	Age         int     `json:"age"`
	Gender      string  `json:"gender"`
	CountryCode string  `json:"country_code"`
	Region      *string `json:"region"`
	HasPhoto    bool    `json:"has_photo"`
	PhotoURL    *string `json:"photo_url"` // <--- 1. AÑADIDO AL JSON DE RESPUESTA
	CreatedAt   string  `json:"created_at"`
}

type listResponse struct {
	Items      []itemResponse `json:"items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	Total      int            `json:"total"`
	TotalPages int            `json:"total_pages"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}
	page, pageSize, ok := parsePaging(w, r)
	if !ok {
		return
	}
	result, err := h.svc.List(r.Context(), userID, page, pageSize)
	if err != nil {
		writeError(w, err)
		return
	}
	items := make([]itemResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, itemResponse{
			EventType:   item.EventType,
			ProfileID:   item.ProfileID.String(),
			DisplayName: item.DisplayName,
			Age:         item.Age,
			Gender:      string(item.Gender),
			CountryCode: item.CountryCode,
			Region:      item.Region,
			HasPhoto:    item.HasPhoto,
			PhotoURL:    item.PhotoURL, // <--- 2. MAPEADO AQUÍ
			CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		})
	}
	httpx.WriteJSON(w, http.StatusOK, listResponse{
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func parsePaging(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	page, pageSize := 1, DefaultPageSize
	for key, target := range map[string]*int{"page": &page, "page_size": &pageSize} {
		if value := r.URL.Query().Get(key); value != "" {
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				httpx.WriteError(w, http.StatusBadRequest, "invalid_param", key+" debe ser un entero >= 1.")
				return 0, 0, false
			}
			*target = n
		}
	}
	return page, pageSize, true
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, profiles.ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Perfil no encontrado.")
		return
	}
	httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.")
}