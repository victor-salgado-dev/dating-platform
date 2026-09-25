package visits

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/httpx"
	"dating-platform/backend/internal/profiles"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type profileResponse struct {
	ProfileID        string  `json:"profile_id"`
	DisplayName      string  `json:"display_name"`
	Age              int     `json:"age"`
	Gender           string  `json:"gender"`
	CountryCode      string  `json:"country_code"`
	Region           *string `json:"region"`
	RelationshipGoal *string `json:"relationship_goal"`
	HasPhoto         bool    `json:"has_photo"`
	VisitedAt        string  `json:"visited_at"`
	// PhotoID es el id de la foto principal, resuelto por el repositorio en
	// la misma consulta que la lista (evita el N+1). Nil si el perfil no
	// tiene fotos.
	PhotoID *string `json:"photo_id"`
	// PhotoURL es la URL de la foto principal ya armada, con el mismo formato
	// que discover/search, likes y activity:
	// /api/v1/profiles/{profileID}/photos/{photoID}/file
	PhotoURL *string `json:"photo_url"`
}

type listResponse struct {
	Items      []profileResponse `json:"items"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	Total      int               `json:"total"`
	TotalPages int               `json:"total_pages"`
}

func (h *Handler) Record(w http.ResponseWriter, r *http.Request) {
	userID, profileID, ok := h.authAndProfileID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Record(r.Context(), userID, profileID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListSent(w http.ResponseWriter, r *http.Request)     { h.list(w, r, "sent") }
func (h *Handler) ListReceived(w http.ResponseWriter, r *http.Request) { h.list(w, r, "received") }
func (h *Handler) ListMutual(w http.ResponseWriter, r *http.Request)   { h.list(w, r, "mutual") }

func (h *Handler) list(w http.ResponseWriter, r *http.Request, mode string) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "unauthenticated", "Inicia sesión para continuar.")
		return
	}
	page, pageSize, ok := parsePaging(w, r)
	if !ok {
		return
	}
	var result *ListResult
	var err error
	switch mode {
	case "sent":
		result, err = h.svc.ListSent(r.Context(), userID, page, pageSize)
	case "received":
		result, err = h.svc.ListReceived(r.Context(), userID, page, pageSize)
	default:
		result, err = h.svc.ListMutual(r.Context(), userID, page, pageSize)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toListResponse(result))
}

func (h *Handler) authAndProfileID(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "unauthenticated", "Inicia sesión para continuar.")
		return uuid.Nil, uuid.Nil, false
	}
	id, err := uuid.Parse(r.PathValue("profileID"))
	if err != nil {
		httpx.WriteError(w, 400, "invalid_id", "ID de perfil inválido.")
		return uuid.Nil, uuid.Nil, false
	}
	return userID, id, true
}

func parsePaging(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	page, size := 1, DefaultPageSize
	for key, target := range map[string]*int{"page": &page, "page_size": &size} {
		if value := r.URL.Query().Get(key); value != "" {
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				httpx.WriteError(w, 400, "invalid_param", key+" debe ser un entero >= 1.")
				return 0, 0, false
			}
			*target = n
		}
	}
	return page, size, true
}

// photoIDString devuelve la representación en string del id de foto, o nil.
func photoIDString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	v := id.String()
	return &v
}

// photoURL arma la URL de la foto principal del perfil a partir de su
// ProfileID y PhotoID, con el mismo formato que discover/search, likes y
// activity: /api/v1/profiles/{profileID}/photos/{photoID}/file
// Devuelve nil si el perfil no tiene foto.
func photoURL(profileID uuid.UUID, photoID *uuid.UUID) *string {
	if photoID == nil {
		return nil
	}
	u := fmt.Sprintf("/api/v1/profiles/%s/photos/%s/file", profileID.String(), photoID.String())
	return &u
}

func toListResponse(result *ListResult) listResponse {
	items := make([]profileResponse, 0, len(result.Items))
	for _, it := range result.Items {
		var goal *string
		if it.RelationshipGoal != nil {
			v := string(*it.RelationshipGoal)
			goal = &v
		}
		items = append(items, profileResponse{
			ProfileID:        it.ProfileID.String(),
			DisplayName:      it.DisplayName,
			Age:              it.Age,
			Gender:           string(it.Gender),
			CountryCode:      it.CountryCode,
			Region:           it.Region,
			RelationshipGoal: goal,
			HasPhoto:         it.HasPhoto,
			VisitedAt:        it.VisitedAt.Format(time.RFC3339),
			PhotoID:          photoIDString(it.PhotoID),
			PhotoURL:         photoURL(it.ProfileID, it.PhotoID),
		})
	}
	return listResponse{
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	}
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrCannotVisitSelf):
		httpx.WriteError(w, 400, "cannot_visit_self", "No puedes registrar visita a tu propio perfil.")
	case errors.Is(err, profiles.ErrNotFound):
		httpx.WriteError(w, 404, "profile_not_found", "Perfil no encontrado.")
	default:
		httpx.WriteError(w, 500, "internal_error", "No se pudo completar la operación.")
	}
}
