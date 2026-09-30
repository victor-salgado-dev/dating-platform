package search

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// popularityLockID es la clave del advisory lock que evita que varias
// instancias del backend refresquen la vista a la vez.
const popularityLockID int64 = 727_001_001

// DefaultPopularityRefreshInterval es cada cuánto se recalcula la
// popularidad. Es un ranking, no un contador en tiempo real.
const DefaultPopularityRefreshInterval = 5 * time.Minute

// StartPopularityRefresher refresca la vista materializada
// profile_popularity una vez al arrancar y luego cada `interval`, hasta que
// ctx se cancele. Está pensada para lanzarse en una goroutine desde main.
func StartPopularityRefresher(ctx context.Context, db *pgxpool.Pool, interval time.Duration) {
	if interval <= 0 {
		interval = DefaultPopularityRefreshInterval
	}

	refresh := func() {
		rctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		if err := refreshPopularity(rctx, db); err != nil && ctx.Err() == nil {
			slog.Error("no se pudo refrescar profile_popularity", "error", err)
		}
	}

	refresh()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refresh()
		}
	}
}

func refreshPopularity(ctx context.Context, db *pgxpool.Pool) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	var locked bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, popularityLockID).Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return nil // otra instancia ya lo está refrescando
	}
	defer func() {
		uctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = conn.Exec(uctx, `SELECT pg_advisory_unlock($1)`, popularityLockID)
	}()

	_, err = conn.Exec(ctx, `REFRESH MATERIALIZED VIEW CONCURRENTLY profile_popularity`)
	return err
}
