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
			cookie, err := r.Cookie(cookieName)
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

			// Sin esto, UserIDFromContext (usado por ListUsers/SuspendUser/
			// ReactivateUser para saber quién es el admin que actúa) no
			// encontraría nada: RequireAdmin valida la sesión por su cuenta
			// en vez de delegar en RequireAuth, así que tiene que dejar el
			// contexto en el mismo estado a mano, con la misma clave.
			ctx := auth.ContextWithUserID(r.Context(), u.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
