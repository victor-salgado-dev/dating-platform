// Package health expone los endpoints de comprobación de estado de la
// aplicación: liveness (¿el proceso está vivo?) y readiness (¿puede
// atender tráfico realmente, incluyendo sus dependencias?).
package health

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"dating-platform/backend/internal/httpx"
)

type Handler struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

func NewHandler(db *pgxpool.Pool, rdb *redis.Client) *Handler {
	return &Handler{DB: db, Redis: rdb}
}

// Liveness responde 200 si el proceso Go está en marcha, sin comprobar
// dependencias externas. Pensado para el healthcheck del propio contenedor.
func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness comprueba PostgreSQL y Redis y responde 200 solo si ambos
// están disponibles. Se expone en /api/v1/health.
func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}
	healthy := true

	if err := h.DB.Ping(ctx); err != nil {
		checks["postgres"] = "error"
		healthy = false
	} else {
		checks["postgres"] = "ok"
	}

	if err := h.Redis.Ping(ctx).Err(); err != nil {
		checks["redis"] = "error"
		healthy = false
	} else {
		checks["redis"] = "ok"
	}

	writeJSONStatus := http.StatusOK
	overall := "ok"
	if !healthy {
		writeJSONStatus = http.StatusServiceUnavailable
		overall = "degraded"
	}

	httpx.WriteJSON(w, writeJSONStatus, map[string]any{
		"status": overall,
		"checks": checks,
	})
}
