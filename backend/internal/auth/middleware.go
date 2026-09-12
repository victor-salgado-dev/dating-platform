package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"dating-platform/backend/internal/httpx"
)

type contextKey string

const userIDContextKey contextKey = "auth_user_id"

// RequireAuth exige una cookie de sesión válida. Si es válida, añade el
// ID del usuario al contexto de la petición (ver UserIDFromContext) y
// continúa; si no, responde 401 sin llegar al handler protegido.
func RequireAuth(svc *Service, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil || cookie.Value == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
				return
			}

			u, err := svc.CurrentUser(r.Context(), cookie.Value)
			if err != nil {
				switch {
				case errors.Is(err, ErrTokenInvalid):
					httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Tu sesión ha caducado.")
				case errors.Is(err, ErrAccountSuspended):
					httpx.WriteError(w, http.StatusForbidden, "account_suspended", "Esta cuenta está suspendida.")
				default:
					httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo verificar la sesión.")
				}
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, u.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext recupera el ID del usuario autenticado, colocado por
// RequireAuth. El segundo valor es false si no hay usuario en contexto.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return v, ok
}
