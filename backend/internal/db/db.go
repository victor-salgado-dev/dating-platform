// Package db gestiona la conexión a PostgreSQL, que es la fuente
// principal de verdad de toda la plataforma.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"dating-platform/backend/internal/config"
)

// NewPool crea un pool de conexiones a PostgreSQL a partir de la
// configuración y verifica que la conexión funciona (ping).
func NewPool(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("db: config inválida: %w", err)
	}

	poolCfg.MaxConns = 10
	poolCfg.MinConns = 1
	poolCfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("db: no se pudo crear el pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping fallido: %w", err)
	}

	return pool, nil
}
