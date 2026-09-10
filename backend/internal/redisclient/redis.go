// Package redisclient gestiona la conexión a Redis. Redis se usa
// exclusivamente para cache, rate limiting, presencia y colas/tareas
// asíncronas: nunca como almacén permanente de datos de dominio.
package redisclient

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"dating-platform/backend/internal/config"
)

// New crea un cliente Redis y verifica la conexión mediante PING.
func New(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping fallido: %w", err)
	}

	return client, nil
}
