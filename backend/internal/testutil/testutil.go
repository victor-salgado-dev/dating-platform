// Package testutil contiene helpers para los tests de INTEGRACIÓN
// (build tag "integration"): conectan a la misma PostgreSQL/Redis que
// usa la app en desarrollo (vía la configuración normal, config.Load),
// así que solo hace falta `docker compose up -d postgres redis` antes
// de ejecutarlos. No se usa en los tests unitarios, que no tocan red.
package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"dating-platform/backend/internal/config"
	"dating-platform/backend/internal/db"
	"dating-platform/backend/internal/redisclient"
)

// RequireDB conecta a PostgreSQL y hace saltar (Skip) el test si no
// está disponible, en vez de fallarlo: así `go test ./...` normal
// (sin -tags=integration) ni siquiera compila estos tests, pero si
// alguien los ejecuta a mano sin la infraestructura levantada, el
// resultado es "SKIP", no un rojo confuso.
func RequireDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	cfg := config.Load()
	pool, err := db.NewPool(context.Background(), cfg.Postgres)
	if err != nil {
		t.Skipf("test de integración omitido: no se pudo conectar a PostgreSQL (¿está `docker compose up -d postgres`?): %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// RequireRedis conecta a Redis, con el mismo criterio que RequireDB.
func RequireRedis(t *testing.T) *redis.Client {
	t.Helper()

	cfg := config.Load()
	rdb, err := redisclient.New(context.Background(), cfg.Redis)
	if err != nil {
		t.Skipf("test de integración omitido: no se pudo conectar a Redis (¿está `docker compose up -d redis`?): %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

// UniqueEmail genera un email de prueba que no choca ni con otras
// ejecuciones de test ni con datos ya existentes en la base de datos.
// El TLD .invalid está reservado por RFC 2606 precisamente para esto.
func UniqueEmail() string {
	return fmt.Sprintf("test-%s@example.invalid", uuid.NewString())
}
