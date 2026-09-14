package contact

import (
	"errors"
	"net/http"

	"dating-platform/backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type sendRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	var req sendRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	if err := h.svc.Send(r.Context(), req.Name, req.Email, req.Message); err != nil {
		var valErr *ValidationError
		if errors.As(err, &valErr) {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_field", valErr.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo enviar el mensaje. Inténtalo de nuevo más tarde.")
		return
	}

	httpx.WriteJSON(w, http.StatusAccepted, map[string]string{"status": "sent"})
}
