package profiles

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time { return c.t }

func newTestCache() (*interestCatalogCache, *fakeClock) {
	clk := &fakeClock{t: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	return &interestCatalogCache{now: clk.now}, clk
}

var testDefs = []InterestDefinition{
	{Key: "cocina", HasLevel: true},
	{Key: "guitarra", HasLevel: false},
}

func TestInterestCatalogCache_LoadsOnceWithinTTL(t *testing.T) {
	c, clk := newTestCache()
	calls := 0
	load := func(context.Context) ([]InterestDefinition, error) { calls++; return testDefs, nil }

	for i := 0; i < 5; i++ {
		snap, err := c.get(context.Background(), load)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if len(snap.list) != 2 || !snap.byKey["cocina"].HasLevel || snap.byKey["guitarra"].HasLevel {
			t.Fatalf("snapshot inesperado: %+v", snap)
		}
		clk.t = clk.t.Add(interestCatalogTTL / 10)
	}
	if calls != 1 {
		t.Errorf("load se llamó %d veces dentro del TTL, se esperaba 1", calls)
	}
}

func TestInterestCatalogCache_ReloadsAfterTTL(t *testing.T) {
	c, clk := newTestCache()
	calls := 0
	load := func(context.Context) ([]InterestDefinition, error) { calls++; return testDefs, nil }

	if _, err := c.get(context.Background(), load); err != nil {
		t.Fatal(err)
	}
	clk.t = clk.t.Add(interestCatalogTTL + time.Second)
	if _, err := c.get(context.Background(), load); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Errorf("load se llamó %d veces, se esperaba 2 tras caducar", calls)
	}
}

func TestInterestCatalogCache_ServesStaleOnReloadError(t *testing.T) {
	c, clk := newTestCache()
	ok := func(context.Context) ([]InterestDefinition, error) { return testDefs, nil }
	boom := func(context.Context) ([]InterestDefinition, error) { return nil, errors.New("bd caída") }

	if _, err := c.get(context.Background(), ok); err != nil {
		t.Fatal(err)
	}
	clk.t = clk.t.Add(interestCatalogTTL + time.Second)

	snap, err := c.get(context.Background(), boom)
	if err != nil {
		t.Fatalf("con copia anterior no debería fallar: %v", err)
	}
	if len(snap.list) != 2 {
		t.Errorf("se esperaba la copia anterior, se obtuvo %+v", snap)
	}
}

func TestInterestCatalogCache_ErrorWithoutPreviousCopy(t *testing.T) {
	c, _ := newTestCache()
	boom := func(context.Context) ([]InterestDefinition, error) { return nil, errors.New("bd caída") }

	if _, err := c.get(context.Background(), boom); err == nil {
		t.Fatal("sin copia anterior se esperaba el error de la carga")
	}
}
