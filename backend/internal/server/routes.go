package server

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/health"
	"dating-platform/backend/internal/profiles"
	"dating-platform/backend/internal/search"
)

// Dependencies agrupa todo lo que necesita el router para construir las
// rutas de los distintos módulos de dominio. Se amplía a medida que se
// añaden módulos en fases futuras (messages, ...), evitando que
// NewRouter acumule un parámetro por dependencia.
type Dependencies struct {
	DB    *pgxpool.Pool
	Redis *redis.Client

	AuthService   *auth.Service
	AuthHandler   *auth.Handler
	SessionCookie string

	ProfilesHandler *profiles.Handler
	SearchHandler   *search.Handler
}

// NewRouter construye el árbol de rutas de la aplicación.
//
// Convención de la API: todos los endpoints de dominio se sirven bajo
// /api/v1/... para permitir versionar la API en el futuro sin romper
// clientes existentes. Los endpoints de infraestructura (healthz) quedan
// fuera del versionado porque no forman parte del contrato de la API.
func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()

	healthHandler := health.NewHandler(deps.DB, deps.Redis)

	// Liveness: usado por el healthcheck del contenedor Docker.
	mux.HandleFunc("GET /healthz", healthHandler.Liveness)

	// Readiness: comprueba dependencias (Postgres, Redis).
	mux.HandleFunc("GET /api/v1/health", healthHandler.Readiness)

	requireAuth := auth.RequireAuth(deps.AuthService, deps.SessionCookie)

	// --- Auth (Fase 3) -----------------------------------------------
	mux.HandleFunc("POST /api/v1/auth/register", deps.AuthHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", deps.AuthHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", deps.AuthHandler.Logout)
	mux.HandleFunc("POST /api/v1/auth/password/forgot", deps.AuthHandler.ForgotPassword)
	mux.HandleFunc("POST /api/v1/auth/password/reset", deps.AuthHandler.ResetPassword)
	mux.HandleFunc("POST /api/v1/auth/email/verify", deps.AuthHandler.VerifyEmail)

	mux.Handle("GET /api/v1/auth/me", requireAuth(http.HandlerFunc(deps.AuthHandler.Me)))
	mux.Handle("POST /api/v1/auth/email/resend", requireAuth(http.HandlerFunc(deps.AuthHandler.ResendVerification)))
	mux.Handle("DELETE /api/v1/auth/account", requireAuth(http.HandlerFunc(deps.AuthHandler.DeleteAccount)))

	// --- Profiles (Fase 4) --------------------------------------------
	// Todo bajo /profiles/me: en V1 solo se gestiona el propio perfil.
	// Ver perfiles de otras personas es Fase 6 (Perfiles públicos).
	mux.Handle("GET /api/v1/profiles/me", requireAuth(http.HandlerFunc(deps.ProfilesHandler.Get)))
	mux.Handle("POST /api/v1/profiles/me", requireAuth(http.HandlerFunc(deps.ProfilesHandler.Create)))
	mux.Handle("PATCH /api/v1/profiles/me", requireAuth(http.HandlerFunc(deps.ProfilesHandler.Update)))

	mux.Handle("POST /api/v1/profiles/me/photos", requireAuth(http.HandlerFunc(deps.ProfilesHandler.UploadPhoto)))
	mux.Handle("GET /api/v1/profiles/me/photos", requireAuth(http.HandlerFunc(deps.ProfilesHandler.ListPhotos)))
	mux.Handle("GET /api/v1/profiles/me/photos/{id}/file", requireAuth(http.HandlerFunc(deps.ProfilesHandler.ServePhoto)))
	mux.Handle("DELETE /api/v1/profiles/me/photos/{id}", requireAuth(http.HandlerFunc(deps.ProfilesHandler.DeletePhoto)))

	// --- Search (Fase 5) ------------------------------------------------
	// Devuelve fichas resumidas (sin bio/intereses/idiomas completos ni
	// fotos servibles entre usuarios): la vista de perfil público
	// detallada, con sus reglas de privacidad, llega en la Fase 6.
	mux.Handle("GET /api/v1/search/profiles", requireAuth(http.HandlerFunc(deps.SearchHandler.Search)))

	var handler http.Handler = mux
	handler = withRecover(handler)
	handler = withLogging(handler)

	return handler
}
