package server

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"dating-platform/backend/internal/admin"
	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/blocking"
	"dating-platform/backend/internal/config"
	"dating-platform/backend/internal/consent"
	"dating-platform/backend/internal/contact"
	"dating-platform/backend/internal/favorites"
	"dating-platform/backend/internal/health"
	"dating-platform/backend/internal/likes"
	"dating-platform/backend/internal/messaging"
	"dating-platform/backend/internal/profiles"
	"dating-platform/backend/internal/ratelimit"
	"dating-platform/backend/internal/reports"
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

	ProfilesHandler  *profiles.Handler
	SearchHandler    *search.Handler
	FavoritesHandler *favorites.Handler
	LikesHandler     *likes.Handler
	MessagingHandler *messaging.Handler
	BlockingHandler  *blocking.Handler
	ReportsHandler   *reports.Handler
	AdminHandler     *admin.Handler
	ConsentHandler   *consent.Handler
	ContactHandler   *contact.Handler

	RateLimiter *ratelimit.Limiter
	Security    config.SecurityConfig
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
	authRateLimit := ratelimit.Middleware(deps.RateLimiter, "auth", deps.Security.AuthRateLimit, deps.Security.AuthRateWindow)

	// --- Auth (Fase 3) -----------------------------------------------
	// Rate limit estricto por IP: son los endpoints más golpeados por
	// fuerza bruta (login), scraping de cuentas (forgot-password) y
	// registro masivo de spam.
	mux.Handle("POST /api/v1/auth/register", authRateLimit(http.HandlerFunc(deps.AuthHandler.Register)))
	mux.Handle("POST /api/v1/auth/login", authRateLimit(http.HandlerFunc(deps.AuthHandler.Login)))
	mux.HandleFunc("POST /api/v1/auth/logout", deps.AuthHandler.Logout)
	mux.Handle("POST /api/v1/auth/password/forgot", authRateLimit(http.HandlerFunc(deps.AuthHandler.ForgotPassword)))
	mux.Handle("POST /api/v1/auth/password/reset", authRateLimit(http.HandlerFunc(deps.AuthHandler.ResetPassword)))
	mux.Handle("POST /api/v1/auth/email/verify", authRateLimit(http.HandlerFunc(deps.AuthHandler.VerifyEmail)))

	mux.Handle("GET /api/v1/auth/me", requireAuth(http.HandlerFunc(deps.AuthHandler.Me)))
	mux.Handle("POST /api/v1/auth/email/resend", authRateLimit(requireAuth(http.HandlerFunc(deps.AuthHandler.ResendVerification))))
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

	// --- Perfiles públicos (Fase 6) -------------------------------------
	// Rutas de solo lectura sobre el perfil de OTRA persona. "me" es un
	// segmento literal y siempre gana sobre {profileID} en las rutas de
	// arriba, así que no hay ambigüedad entre ambos bloques.
	mux.Handle("GET /api/v1/profiles/{profileID}", requireAuth(http.HandlerFunc(deps.ProfilesHandler.GetPublic)))
	mux.Handle("GET /api/v1/profiles/{profileID}/photos", requireAuth(http.HandlerFunc(deps.ProfilesHandler.ListPublicPhotos)))
	mux.Handle("GET /api/v1/profiles/{profileID}/photos/{photoID}/file", requireAuth(http.HandlerFunc(deps.ProfilesHandler.ServePublicPhoto)))

	// --- Search (Fase 5) ------------------------------------------------
	// Devuelve fichas resumidas con el profile_id de cada resultado; la
	// vista de perfil público detallada vive en las rutas de arriba.
	mux.Handle("GET /api/v1/search/profiles", requireAuth(http.HandlerFunc(deps.SearchHandler.Search)))

	// --- Favoritos (Fase 7) ---------------------------------------------
	mux.Handle("GET /api/v1/favorites", requireAuth(http.HandlerFunc(deps.FavoritesHandler.List)))
	mux.Handle("POST /api/v1/favorites/{profileID}", requireAuth(http.HandlerFunc(deps.FavoritesHandler.Add)))
	mux.Handle("DELETE /api/v1/favorites/{profileID}", requireAuth(http.HandlerFunc(deps.FavoritesHandler.Remove)))
	mux.Handle("GET /api/v1/favorites/{profileID}", requireAuth(http.HandlerFunc(deps.FavoritesHandler.Status)))

	// --- Likes y matches (Fase 15) --------------------------------------
	mux.Handle("POST /api/v1/likes/{profileID}", requireAuth(http.HandlerFunc(deps.LikesHandler.Add)))
	mux.Handle("DELETE /api/v1/likes/{profileID}", requireAuth(http.HandlerFunc(deps.LikesHandler.Remove)))
	mux.Handle("GET /api/v1/likes/{profileID}", requireAuth(http.HandlerFunc(deps.LikesHandler.Status)))
	mux.Handle("GET /api/v1/likes/sent", requireAuth(http.HandlerFunc(deps.LikesHandler.ListSent)))
	mux.Handle("GET /api/v1/likes/received", requireAuth(http.HandlerFunc(deps.LikesHandler.ListReceived)))
	mux.Handle("GET /api/v1/matches", requireAuth(http.HandlerFunc(deps.LikesHandler.ListMatches)))
	mux.Handle("GET /api/v1/matches/{profileID}", requireAuth(http.HandlerFunc(deps.LikesHandler.MatchStatus)))

	// --- Mensajería (Fase 8) ---------------------------------------------
	// "to/{profileID}" inicia o continúa la conversación con esa persona
	// (find-or-create); una vez abierta, se puede seguir escribiendo y
	// paginando por conversationID sin volver a pasar por el perfil.
	mux.Handle("POST /api/v1/messages/to/{profileID}", requireAuth(http.HandlerFunc(deps.MessagingHandler.SendToProfile)))
	mux.Handle("GET /api/v1/messages/conversations", requireAuth(http.HandlerFunc(deps.MessagingHandler.ListConversations)))
	mux.Handle("GET /api/v1/messages/conversations/{conversationID}/messages", requireAuth(http.HandlerFunc(deps.MessagingHandler.ListMessages)))
	mux.Handle("POST /api/v1/messages/conversations/{conversationID}/messages", requireAuth(http.HandlerFunc(deps.MessagingHandler.SendInConversation)))

	// --- Bloqueo (Fase 9) ------------------------------------------------
	mux.Handle("GET /api/v1/blocks", requireAuth(http.HandlerFunc(deps.BlockingHandler.List)))
	mux.Handle("POST /api/v1/blocks/{profileID}", requireAuth(http.HandlerFunc(deps.BlockingHandler.Add)))
	mux.Handle("DELETE /api/v1/blocks/{profileID}", requireAuth(http.HandlerFunc(deps.BlockingHandler.Remove)))
	mux.Handle("GET /api/v1/blocks/{profileID}", requireAuth(http.HandlerFunc(deps.BlockingHandler.Status)))

	// --- Reportes (Fase 9) ------------------------------------------------
	// Solo creación aquí; revisarlos es la Fase 10 (panel de moderación).
	mux.Handle("POST /api/v1/reports/{profileID}", requireAuth(http.HandlerFunc(deps.ReportsHandler.Create)))

	// --- Administración (Fase 10) ----------------------------------------
	// Rutas distintas de todo lo anterior: exigen rol admin, no solo
	// sesión, y trabajan con user_id (no profile_id) porque son
	// herramientas internas, no cara al público.
	requireAdmin := admin.RequireAdmin(deps.AuthService, deps.SessionCookie)

	mux.Handle("GET /api/v1/admin/users", requireAdmin(http.HandlerFunc(deps.AdminHandler.ListUsers)))
	mux.Handle("POST /api/v1/admin/users/{userID}/suspend", requireAdmin(http.HandlerFunc(deps.AdminHandler.SuspendUser)))
	mux.Handle("POST /api/v1/admin/users/{userID}/reactivate", requireAdmin(http.HandlerFunc(deps.AdminHandler.ReactivateUser)))
	mux.Handle("GET /api/v1/admin/reports", requireAdmin(http.HandlerFunc(deps.AdminHandler.ListReports)))
	mux.Handle("POST /api/v1/admin/reports/{reportID}/resolve", requireAdmin(http.HandlerFunc(deps.AdminHandler.ResolveReport)))

	// --- Legal y privacidad (Fase 13) ------------------------------------
	// El historial de consentimientos es propio, con sesión normal.
	mux.Handle("GET /api/v1/consents/me", requireAuth(http.HandlerFunc(deps.ConsentHandler.ListMine)))

	// El formulario de contacto es deliberadamente público (alguien sin
	// cuenta también tiene que poder escribir), así que NO pasa por
	// requireAuth — pero sí por el mismo rate limit estricto que auth,
	// para no convertirse en un vector de spam/abuso sin control.
	mux.Handle("POST /api/v1/contact", authRateLimit(http.HandlerFunc(deps.ContactHandler.Send)))

	globalRateLimit := ratelimit.Middleware(deps.RateLimiter, "global", deps.Security.GlobalRateLimit, deps.Security.GlobalRateWindow)

	var handler http.Handler = mux
	handler = globalRateLimit(handler)
	handler = withRecover(handler)
	handler = withLogging(handler)
	handler = withSecurityHeaders(handler)
	handler = withMaxBodySize(handler, deps.Security.MaxRequestBodyBytes)

	return handler
}
