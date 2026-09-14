package auth

import (
	"errors"
	"net/http"
	"time"

	"dating-platform/backend/internal/config"
	"dating-platform/backend/internal/consent"
	"dating-platform/backend/internal/httpx"
	"dating-platform/backend/internal/users"
)

// Handler expone el módulo auth como endpoints HTTP bajo /api/v1/auth.
type Handler struct {
	svc *Service
	cfg config.AuthConfig
}

func NewHandler(svc *Service, cfg config.AuthConfig) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// --- DTOs de petición/respuesta -------------------------------------------

type registerRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	AcceptedTerms bool   `json:"accepted_terms"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

type deleteAccountRequest struct {
	Password string `json:"password"`
}

type userResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

func toUserResponse(u *users.User) userResponse {
	return userResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		EmailVerified: u.IsEmailVerified(),
		Status:        string(u.Status),
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
	}
}

// --- Handlers ---------------------------------------------------------------

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	u, token, err := h.svc.Register(r.Context(), req.Email, req.Password, req.AcceptedTerms)
	if err != nil {
		switch {
		case errors.Is(err, ErrWeakPassword):
			httpx.WriteError(w, http.StatusBadRequest, "weak_password", "La contraseña debe tener al menos 8 caracteres.")
		case errors.Is(err, users.ErrDuplicateEmail):
			httpx.WriteError(w, http.StatusConflict, "email_taken", "Ese email ya está registrado.")
		case errors.Is(err, ErrInvalidEmail):
			httpx.WriteError(w, http.StatusBadRequest, "invalid_email", "El formato del email no es válido.")
		case errors.Is(err, consent.ErrTermsNotAccepted):
			httpx.WriteError(w, http.StatusBadRequest, "terms_not_accepted", "Debes aceptar los Términos y la Política de Privacidad.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar el registro.")
		}
		return
	}

	h.setSessionCookie(w, token)
	httpx.WriteJSON(w, http.StatusCreated, toUserResponse(u))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	u, token, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrAccountSuspended):
			httpx.WriteError(w, http.StatusForbidden, "account_suspended", "Esta cuenta está suspendida.")
		case errors.Is(err, ErrInvalidCredentials):
			httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Email o contraseña incorrectos.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo iniciar sesión.")
		}
		return
	}

	h.setSessionCookie(w, token)
	httpx.WriteJSON(w, http.StatusOK, toUserResponse(u))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.cfg.CookieName)
	if err == nil && cookie.Value != "" {
		_ = h.svc.Logout(r.Context(), cookie.Value)
	}

	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	u, err := h.svc.users.GetByID(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toUserResponse(u))
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	// Siempre 202, exista o no la cuenta: evita enumeración de emails.
	_ = h.svc.RequestPasswordReset(r.Context(), req.Email)
	httpx.WriteJSON(w, http.StatusAccepted, map[string]string{"status": "if_account_exists_email_sent"})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	err := h.svc.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, ErrWeakPassword):
			httpx.WriteError(w, http.StatusBadRequest, "weak_password", "La contraseña debe tener al menos 8 caracteres.")
		case errors.Is(err, ErrTokenInvalid):
			httpx.WriteError(w, http.StatusBadRequest, "invalid_token", "El enlace de restablecimiento no es válido o ha caducado.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo restablecer la contraseña.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	if err := h.svc.VerifyEmail(r.Context(), req.Token); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_token", "El enlace de verificación no es válido o ha caducado.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "verified"})
}

func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	err := h.svc.ResendVerification(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyVerified) {
			httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "already_verified"})
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo reenviar la verificación.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	var req deleteAccountRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	if err := h.svc.DeleteAccount(r.Context(), userID, req.Password); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusForbidden, "invalid_credentials", "Contraseña incorrecta.")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo eliminar la cuenta.")
		return
	}

	if cookie, err := r.Cookie(h.cfg.CookieName); err == nil && cookie.Value != "" {
		_ = h.svc.Logout(r.Context(), cookie.Value)
	}
	h.clearSessionCookie(w)

	w.WriteHeader(http.StatusNoContent)
}

// --- Cookies ------------------------------------------------------------

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.cfg.SessionTTL.Seconds()),
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
