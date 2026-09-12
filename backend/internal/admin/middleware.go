// Package admin implementa el panel de administración (Fase 10):
// listar y suspender cuentas, listar y resolver reportes. Depende de
// users.Repository y reports.Repository (interfaces) y de auth.Service
// para la autenticación de sesión; ninguno de esos módulos depende de
// admin, así que no hay ciclos.
package admin

import (
	"errors"
	"net/http"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/httpx"
)

// RequireAdmin exige una sesión válida (igual que auth.RequireAuth) Y
// que la cuenta tenga rol 'admin'. Deliberadamente no se compone sobre
// auth.RequireAuth: necesitamos el *users.User completo (para mirar el
// rol), no solo el ID que deja auth.RequireAuth en el contexto.
func RequireAdmin(authSvc *auth.Service, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
				return
			}

			u, err := authSvc.CurrentUser(r.Context(), cookie.Value)
			if err != nil {
				switch {
				case errors.Is(err, auth.ErrTokenInvalid):
					httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Tu sesión ha caducado.")
				case errors.Is(err, auth.ErrAccountSuspended):
					httpx.WriteError(w, http.StatusForbidden, "account_suspended", "Esta cuenta está suspendida.")
				default:
					httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo verificar la sesión.")
				}
				return
			}

			if !u.IsAdmin() {
				httpx.WriteError(w, http.StatusForbidden, "forbidden", "No tienes permisos de administración.")
				return
			}

			// Inyectar el ID de usuario en el contexto
ctx := context.WithValue(r.Context(), auth.UserIDContextKey, u.ID)
next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
