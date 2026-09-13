package ratelimit

import (
	"log/slog"
	"net/http"
	"time"

	"dating-platform/backend/internal/httpx"
)

// Middleware aplica un límite de `limit` peticiones por `window` y por
// IP de cliente. `scope` distingue distintos límites en Redis (p. ej.
// "global" vs "auth") para que no compartan contador entre sí.
func Middleware(limiter *Limiter, scope string, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := scope + ":" + httpx.ClientIP(r)

			allowed, err := limiter.Allow(r.Context(), key, limit, window)
			if err != nil {
				// Un fallo de Redis aquí no debe tumbar toda la API: se
				// registra y se deja pasar la petición (fail-open). El
				// rate limiting es una capa de defensa, no la única.
				slog.Error("rate limit: no se pudo comprobar el límite", "error", err, "scope", scope)
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", "60")
				httpx.WriteError(w, http.StatusTooManyRequests, "rate_limited", "Demasiadas peticiones. Inténtalo de nuevo más tarde.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
