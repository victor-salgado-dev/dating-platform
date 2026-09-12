package admin

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/httpx"
	"dating-platform/backend/internal/reports"
	"dating-platform/backend/internal/users"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// --- DTOs -----------------------------------------------------------------

type userResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Status        string  `json:"status"`
	Role          string  `json:"role"`
	EmailVerified bool    `json:"email_verified"`
	CreatedAt     string  `json:"created_at"`
	DeletedAt     *string `json:"deleted_at"`
}

func toUserResponse(u *users.User) userResponse {
	var deletedAt *string
	if u.DeletedAt != nil {
		v := u.DeletedAt.Format(time.RFC3339)
		deletedAt = &v
	}
	return userResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Status:        string(u.Status),
		Role:          string(u.Role),
		EmailVerified: u.IsEmailVerified(),
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
		DeletedAt:     deletedAt,
	}
}

type usersListResponse struct {
	Items      []userResponse `json:"items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	Total      int            `json:"total"`
	TotalPages int            `json:"total_pages"`
}

type reportResponse struct {
	ID           string  `json:"id"`
	ReporterID   string  `json:"reporter_id"`
	ReporterName *string `json:"reporter_name"`
	ReportedID   string  `json:"reported_id"`
	ReportedName *string `json:"reported_name"`
	Reason       string  `json:"reason"`
	Description  *string `json:"description"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

func toReportResponse(it *reports.ListItem) reportResponse {
	return reportResponse{
		ID:           it.ID.String(),
		ReporterID:   it.ReporterID.String(),
		ReporterName: it.ReporterName,
		ReportedID:   it.ReportedID.String(),
		ReportedName: it.ReportedName,
		Reason:       string(it.Reason),
		Description:  it.Description,
		Status:       string(it.Status),
		CreatedAt:    it.CreatedAt.Format(time.RFC3339),
	}
}

type reportsListResponse struct {
	Items      []reportResponse `json:"items"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	Total      int              `json:"total"`
	TotalPages int              `json:"total_pages"`
}

type resolveReportRequest struct {
	Status string `json:"status"`
}

// --- Handlers: usuarios -------------------------------------------------

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, pageSize, err := parsePaging(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_param", err.Error())
		return
	}

	var statusFilter *users.Status
	if v := r.URL.Query().Get("status"); v != "" {
		s := users.Status(v)
		statusFilter = &s
	}

	result, err := h.svc.ListUsers(r.Context(), page, pageSize, statusFilter)
	if err != nil {
		writeAdminError(w, err)
		return
	}

	items := make([]userResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, toUserResponse(&result.Items[i]))
	}

	httpx.WriteJSON(w, http.StatusOK, usersListResponse{
		Items: items, Page: result.Page, PageSize: result.PageSize,
		Total: result.Total, TotalPages: result.TotalPages,
	})
}

func (h *Handler) SuspendUser(w http.ResponseWriter, r *http.Request) {
	adminID, targetID, ok := h.adminAndTargetUserID(w, r)
	if !ok {
		return
	}

	if err := h.svc.SuspendUser(r.Context(), adminID, targetID); err != nil {
		writeAdminError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ReactivateUser(w http.ResponseWriter, r *http.Request) {
	_, targetID, ok := h.adminAndTargetUserID(w, r)
	if !ok {
		return
	}

	if err := h.svc.ReactivateUser(r.Context(), targetID); err != nil {
		writeAdminError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Handlers: reportes ---------------------------------------------------

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	page, pageSize, err := parsePaging(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_param", err.Error())
		return
	}

	var statusFilter *reports.Status
	if v := r.URL.Query().Get("status"); v != "" {
		s := reports.Status(v)
		statusFilter = &s
	}

	result, err := h.svc.ListReports(r.Context(), page, pageSize, statusFilter)
	if err != nil {
		writeAdminError(w, err)
		return
	}

	items := make([]reportResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, toReportResponse(&result.Items[i]))
	}

	httpx.WriteJSON(w, http.StatusOK, reportsListResponse{
		Items: items, Page: result.Page, PageSize: result.PageSize,
		Total: result.Total, TotalPages: result.TotalPages,
	})
}

func (h *Handler) ResolveReport(w http.ResponseWriter, r *http.Request) {
	reportID, err := uuid.Parse(r.PathValue("reportID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de reporte inválido.")
		return
	}

	var req resolveReportRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	if err := h.svc.ResolveReport(r.Context(), reportID, reports.Status(req.Status)); err != nil {
		writeAdminError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Helpers ----------------------------------------------------------

func (h *Handler) adminAndTargetUserID(w http.ResponseWriter, r *http.Request) (adminID, targetID uuid.UUID, ok bool) {
	adminID, authOK := auth.UserIDFromContext(r.Context())
	if !authOK {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return uuid.Nil, uuid.Nil, false
	}

	targetID, err := uuid.Parse(r.PathValue("userID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de usuario inválido.")
		return uuid.Nil, uuid.Nil, false
	}

	return adminID, targetID, true
}

func parsePaging(r *http.Request) (page, pageSize int, err error) {
	page = 1
	if v := r.URL.Query().Get("page"); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil || n < 1 {
			return 0, 0, errors.New("page debe ser un entero >= 1")
		}
		page = n
	}

	pageSize = 20
	if v := r.URL.Query().Get("page_size"); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil || n < 1 {
			return 0, 0, errors.New("page_size debe ser un entero >= 1")
		}
		pageSize = n
	}

	return page, pageSize, nil
}

func writeAdminError(w http.ResponseWriter, err error) {
	var valErr *ValidationError
	switch {
	case errors.As(err, &valErr):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_field", valErr.Error())
	case errors.Is(err, ErrCannotActOnSelf):
		httpx.WriteError(w, http.StatusBadRequest, "cannot_act_on_self", "No puedes realizar esta acción sobre tu propia cuenta.")
	case errors.Is(err, users.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "user_not_found", "Usuario no encontrado.")
	case errors.Is(err, reports.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "report_not_found", "Reporte no encontrado.")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.")
	}
}
