// Package ratelimit implementa limitación de peticiones respaldada por
// Redis (sección 4: Redis se usa para rate limiting, nunca como fuente
// de verdad). El algoritmo es "ventana fija": simple, barato, y de
// sobra para el volumen de V1; una ventana deslizante o un token
// bucket son mejoras posibles si el tráfico real lo pidiera, no antes.
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Limiter {
	return &Limiter{rdb: rdb}
}

// Allow indica si una nueva petición bajo `key` cabe dentro de `limit`
// peticiones por `window`. Cada key lleva su propio contador en Redis
// con expiración igual a la ventana: pasado ese tiempo, se reinicia.
func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	redisKey := "ratelimit:" + key

	count, err := l.rdb.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, fmt.Errorf("ratelimit: incrementar contador: %w", err)
	}

	if count == 1 {
		// Primera petición de la ventana: fija cuándo expira el contador.
		// Si esto fallara, en el peor caso la key no expira y el
		// contador se queda alto más tiempo del debido (falla cerrado
		// hacia "más estricto", no hacia "sin límite").
		if err := l.rdb.Expire(ctx, redisKey, window).Err(); err != nil {
			return false, fmt.Errorf("ratelimit: fijar expiración: %w", err)
		}
	}

	return count <= int64(limit), nil
}
