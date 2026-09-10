package server

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"dating-platform/backend/internal/health"
)

// NewRouter construye el árbol de rutas de la aplicación.
//
// Convención de la API: todos los endpoints de dominio se sirven bajo
// /api/v1/... para permitir versionar la API en el futuro sin romper
// clientes existentes. Los endpoints de infraestructura (healthz) quedan
// fuera del versionado porque no forman parte del contrato de la API.
func NewRouter(db *pgxpool.Pool, rdb *redis.Client) http.Handler {
	mux := http.NewServeMux()

	healthHandler := health.NewHandler(db, rdb)

	// Liveness: usado por el healthcheck del contenedor Docker.
	mux.HandleFunc("GET /healthz", healthHandler.Liveness)

	// Readiness: comprueba dependencias (Postgres, Redis).
	mux.HandleFunc("GET /api/v1/health", healthHandler.Readiness)

	var handler http.Handler = mux
	handler = withRecover(handler)
	handler = withLogging(handler)

	return handler
}
