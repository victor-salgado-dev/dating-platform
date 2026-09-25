package messaging

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

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// --- DTOs -----------------------------------------------------------------

type sendMessageRequest struct {
	Body string `json:"body"`
}

type messageResponse struct {
	ID             string  `json:"id"`
	ConversationID string  `json:"conversation_id"`
	Body           string  `json:"body"`
	CreatedAt      string  `json:"created_at"`
	IsMine         bool    `json:"is_mine"`
	ReadAt         *string `json:"read_at"`
}

func toMessageResponse(m *Message, requesterID uuid.UUID) messageResponse {
	var readAt *string
	if m.ReadAt != nil {
		v := m.ReadAt.Format(time.RFC3339)
		readAt = &v
	}
	return messageResponse{
		ID:             m.ID.String(),
		ConversationID: m.ConversationID.String(),
		Body:           m.Body,
		CreatedAt:      m.CreatedAt.Format(time.RFC3339),
		IsMine:         m.SenderID == requesterID,
		ReadAt:         readAt,
	}
}

type participantResponse struct {
	ProfileID   string  `json:"profile_id"`
	DisplayName string  `json:"display_name"`
	Age         int     `json:"age"`
	Gender      string  `json:"gender"`
	CountryCode string  `json:"country_code"`
	Region      *string `json:"region"`
	HasPhoto    bool    `json:"has_photo"`
	// PhotoID es el id de la foto principal, resuelto por el repositorio en
	// la misma consulta que la lista (evita el N+1). Nil si el perfil no
	// tiene fotos.
	PhotoID *string `json:"photo_id"`
	// PhotoURL es la URL de la foto principal ya armada, con el mismo formato
	// que discover/search, likes, visits, activity y favorites:
	// /api/v1/profiles/{profileID}/photos/{photoID}/file
	PhotoURL *string `json:"photo_url"`
}

type lastMessageResponse struct {
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	IsMine    bool   `json:"is_mine"`
}

type conversationResponse struct {
	ConversationID   string              `json:"conversation_id"`
	OtherParticipant participantResponse `json:"other_participant"`
	LastMessage      lastMessageResponse `json:"last_message"`
	UnreadCount      int                 `json:"unread_count"`
}

func toConversationResponse(c *ConversationSummary) conversationResponse {
	return conversationResponse{
		ConversationID: c.ConversationID.String(),
		OtherParticipant: participantResponse{
			ProfileID:   c.OtherParticipant.ProfileID.String(),
			DisplayName: c.OtherParticipant.DisplayName,
			Age:         c.OtherParticipant.Age,
			Gender:      string(c.OtherParticipant.Gender),
			CountryCode: c.OtherParticipant.CountryCode,
			Region:      c.OtherParticipant.Region,
			HasPhoto:    c.OtherParticipant.HasPhoto,
			PhotoID:     photoIDString(c.OtherParticipant.PhotoID),
			PhotoURL:    photoURL(c.OtherParticipant.ProfileID, c.OtherParticipant.PhotoID),
		},
		LastMessage: lastMessageResponse{
			Body:      c.LastMessageBody,
			CreatedAt: c.LastMessageAt.Format(time.RFC3339),
			IsMine:    c.LastMessageIsMine,
		},
		UnreadCount: c.UnreadCount,
	}
}

type conversationListResponse struct {
	Items      []conversationResponse `json:"items"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	Total      int                    `json:"total"`
	TotalPages int                    `json:"total_pages"`
}

type messageListResponse struct {
	ConversationID string            `json:"conversation_id"`
	Items          []messageResponse `json:"items"`
	Page           int               `json:"page"`
	PageSize       int               `json:"page_size"`
	Total          int               `json:"total"`
	TotalPages     int               `json:"total_pages"`
}

// --- Handlers ---------------------------------------------------------

func (h *Handler) SendToProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	profileID, err := uuid.Parse(r.PathValue("profileID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de perfil inválido.")
		return
	}

	var req sendMessageRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	msg, err := h.svc.SendToProfile(r.Context(), userID, profileID, req.Body)
	if err != nil {
		writeMessagingError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toMessageResponse(msg, userID))
}

func (h *Handler) SendInConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	conversationID, err := uuid.Parse(r.PathValue("conversationID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de conversación inválido.")
		return
	}

	var req sendMessageRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	msg, err := h.svc.SendInConversation(r.Context(), userID, conversationID, req.Body)
	if err != nil {
		writeMessagingError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toMessageResponse(msg, userID))
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	conversationID, err := uuid.Parse(r.PathValue("conversationID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de conversación inválido.")
		return
	}

	page, pageSize, perr := parsePaging(r)
	if perr != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_param", perr.Error())
		return
	}

	result, err := h.svc.ListMessages(r.Context(), userID, conversationID, page, pageSize)
	if err != nil {
		writeMessagingError(w, err)
		return
	}

	items := make([]messageResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, toMessageResponse(&result.Items[i], userID))
	}

	httpx.WriteJSON(w, http.StatusOK, messageListResponse{
		ConversationID: result.ConversationID.String(),
		Items:          items,
		Page:           result.Page,
		PageSize:       result.PageSize,
		Total:          result.Total,
		TotalPages:     result.TotalPages,
	})
}

func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	page, pageSize, perr := parsePaging(r)
	if perr != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_param", perr.Error())
		return
	}

	result, err := h.svc.ListConversations(r.Context(), userID, page, pageSize)
	if err != nil {
		writeMessagingError(w, err)
		return
	}

	items := make([]conversationResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, toConversationResponse(&result.Items[i]))
	}

	httpx.WriteJSON(w, http.StatusOK, conversationListResponse{
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

// --- Helpers ----------------------------------------------------------

func parsePaging(r *http.Request) (page, pageSize int, err error) {
	page = 1
	if v := r.URL.Query().Get("page"); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil || n < 1 {
			return 0, 0, errors.New("page debe ser un entero >= 1")
		}
		page = n
	}

	pageSize = DefaultPageSize
	if v := r.URL.Query().Get("page_size"); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil || n < 1 {
			return 0, 0, errors.New("page_size debe ser un entero >= 1")
		}
		pageSize = n
	}

	return page, pageSize, nil
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
// ProfileID y PhotoID, con el mismo formato que discover/search, likes,
// visits, activity y favorites:
// /api/v1/profiles/{profileID}/photos/{photoID}/file
// Devuelve nil si el perfil no tiene foto.
func photoURL(profileID uuid.UUID, photoID *uuid.UUID) *string {
	if photoID == nil {
		return nil
	}
	u := fmt.Sprintf("/api/v1/profiles/%s/photos/%s/file", profileID.String(), photoID.String())
	return &u
}

func writeMessagingError(w http.ResponseWriter, err error) {
	var valErr *ValidationError
	switch {
	case errors.As(err, &valErr):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_field", valErr.Error())
	case errors.Is(err, ErrCannotMessageSelf):
		httpx.WriteError(w, http.StatusBadRequest, "cannot_message_self", "No puedes enviarte un mensaje a ti mismo.")
	case errors.Is(err, ErrConversationNotFound):
		httpx.WriteError(w, http.StatusNotFound, "conversation_not_found", "Conversación no encontrada.")
	case errors.Is(err, profiles.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Perfil no encontrado.")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.")
	}
}
