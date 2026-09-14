//go:build integration

package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/ratelimit"
	"dating-platform/backend/internal/testutil"
)

func TestIntegration_LimiterAllowsUpToLimitThenBlocks(t *testing.T) {
	rdb := testutil.RequireRedis(t)
	limiter := ratelimit.New(rdb)
	ctx := context.Background()

	// Key única por ejecución: no debe chocar con otros tests ni con
	// ejecuciones anteriores que dejaran un contador vivo.
	key := "test:" + uuid.NewString()

	const limit = 3
	window := time.Minute

	for i := 1; i <= limit; i++ {
		allowed, err := limiter.Allow(ctx, key, limit, window)
		if err != nil {
			t.Fatalf("Allow (petición %d): %v", i, err)
		}
		if !allowed {
			t.Errorf("la petición %d debería estar permitida (límite = %d)", i, limit)
		}
	}

	// La petición limit+1 debe bloquearse.
	allowed, err := limiter.Allow(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("Allow (petición %d): %v", limit+1, err)
	}
	if allowed {
		t.Errorf("la petición %d debería bloquearse (supera el límite de %d)", limit+1, limit)
	}
}

func TestIntegration_LimiterUsesIndependentKeys(t *testing.T) {
	rdb := testutil.RequireRedis(t)
	limiter := ratelimit.New(rdb)
	ctx := context.Background()

	keyA := "test:" + uuid.NewString()
	keyB := "test:" + uuid.NewString()

	// Agota el límite de A.
	if _, err := limiter.Allow(ctx, keyA, 1, time.Minute); err != nil {
		t.Fatalf("Allow keyA: %v", err)
	}
	allowedA, err := limiter.Allow(ctx, keyA, 1, time.Minute)
	if err != nil {
		t.Fatalf("Allow keyA (2ª vez): %v", err)
	}
	if allowedA {
		t.Error("keyA debería estar bloqueada tras superar su límite")
	}

	// B no debería verse afectada por el límite de A.
	allowedB, err := limiter.Allow(ctx, keyB, 1, time.Minute)
	if err != nil {
		t.Fatalf("Allow keyB: %v", err)
	}
	if !allowedB {
		t.Error("keyB tiene su propio contador independiente y debería estar permitida")
	}
}
