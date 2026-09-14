package consent

import (
	"net/http"
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

type consentResponse struct {
	DocumentType    string `json:"document_type"`
	DocumentVersion string `json:"document_version"`
	AcceptedAt      string `json:"accepted_at"`
}

// ListMine devuelve el historial de consentimientos del usuario
// autenticado: transparencia sobre qué aceptó y cuándo.
func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	items, err := h.svc.ListMyConsents(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo obtener el historial de consentimientos.")
		return
	}

	resp := make([]consentResponse, 0, len(items))
	for _, c := range items {
		resp = append(resp, consentResponse{
			DocumentType:    string(c.DocumentType),
			DocumentVersion: c.DocumentVersion,
			AcceptedAt:      c.AcceptedAt.Format(time.RFC3339),
		})
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}
