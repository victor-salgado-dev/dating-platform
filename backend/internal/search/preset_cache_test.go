package search

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

func testEntry(n int) (*presetEntry, []uuid.UUID) {
	e := &presetEntry{complete: true, members: map[uuid.UUID]struct{}{}}
	ids := make([]uuid.UUID, n)
	for i := 0; i < n; i++ {
		ids[i] = uuid.New()
		e.cards = append(e.cards, presetCard{ProfileID: ids[i], DisplayName: string(rune('A' + i))})
		e.members[ids[i]] = struct{}{}
	}
	return e, ids
}

func names(cards []presetCard) string {
	s := ""
	for _, c := range cards {
		s += c.DisplayName
	}
	return s
}

func TestPageOfCards(t *testing.T) {
	e, ids := testEntry(7) // A B C D E F G

	tests := []struct {
		name      string
		excluded  []uuid.UUID
		page, siz int
		want      string
		wantTotal int
	}{
		{"sin exclusiones, página 1", nil, 1, 3, "ABC", 7},
		{"sin exclusiones, última página parcial", nil, 3, 3, "G", 7},
		{"se descuenta un excluido de la primera página", []uuid.UUID{ids[1]}, 1, 3, "ACD", 6},
		{"la segunda página sigue sin huecos ni repetidos", []uuid.UUID{ids[1]}, 2, 3, "EFG", 6},
		{"excluido que no está en la lista no cuenta", []uuid.UUID{uuid.New()}, 1, 3, "ABC", 7},
		{"varios excluidos", []uuid.UUID{ids[0], ids[2], ids[6]}, 1, 10, "BDEF", 4},
		{"página fuera de rango conserva el total", nil, 9, 3, "", 7},
		{"todos excluidos", ids, 1, 3, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards, total := pageOfCards(e, tt.excluded, ViewerScope{}, time.Now(), tt.page, tt.siz)
			if got := names(cards); got != tt.want {
				t.Errorf("tarjetas = %q, quería %q", got, tt.want)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, quería %d", total, tt.wantTotal)
			}
		})
	}
}

func TestIsPlainPreset(t *testing.T) {
	if !isPlainPreset(Params{Sort: SortRecent}) {
		t.Error("new-members sin filtros debería ser cacheable")
	}
	if !isPlainPreset(Params{Sort: SortPopular}) {
		t.Error("popular sin filtros debería ser cacheable")
	}
	if isPlainPreset(Params{Sort: SortAgeAsc}) {
		t.Error("otros órdenes no se cachean")
	}

	country := "ES"
	if isPlainPreset(Params{Sort: SortRecent, Filters: Filters{CountryCode: &country}}) {
		t.Error("con filtro de país no debe usar la caché")
	}
	if isPlainPreset(Params{Sort: SortRecent, Filters: Filters{OnlineNow: true}}) {
		t.Error("online-now no debe usar la caché")
	}
	if isPlainPreset(Params{Sort: SortRecent, Filters: Filters{Sports: []string{"x"}}}) {
		t.Error("con filtros de lista no debe usar la caché")
	}
}

func TestPresetCacheBuildsOnceAndReuses(t *testing.T) {
	c := newPresetCache(time.Hour)
	var builds atomic.Int32
	build := func(context.Context) (*presetEntry, error) {
		builds.Add(1)
		time.Sleep(20 * time.Millisecond)
		return &presetEntry{complete: true}, nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.get(context.Background(), SortRecent, build); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	if n := builds.Load(); n != 1 {
		t.Errorf("con 50 peticiones simultáneas se construyó %d veces, quería 1", n)
	}
}

func TestPresetCacheServesStaleWhileRebuilding(t *testing.T) {
	c := newPresetCache(time.Millisecond)

	first := &presetEntry{complete: true}
	if _, err := c.get(context.Background(), SortRecent, func(context.Context) (*presetEntry, error) { return first, nil }); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond) // caduca

	release := make(chan struct{})
	started := make(chan struct{})
	go func() {
		_, _ = c.get(context.Background(), SortRecent, func(context.Context) (*presetEntry, error) {
			close(started)
			<-release
			return &presetEntry{complete: true}, nil
		})
	}()
	<-started

	// Mientras la otra petición reconstruye, ésta no espera: recibe la anterior.
	got, err := c.get(context.Background(), SortRecent, func(context.Context) (*presetEntry, error) {
		t.Error("no debería lanzar una segunda reconstrucción")
		return nil, errors.New("no")
	})
	if err != nil || got != first {
		t.Errorf("debería servir la versión anterior, got=%v err=%v", got, err)
	}
	close(release)
}

func TestPresetCacheKeepsServingOnBuildError(t *testing.T) {
	c := newPresetCache(time.Millisecond)

	first := &presetEntry{complete: true}
	_, _ = c.get(context.Background(), SortRecent, func(context.Context) (*presetEntry, error) { return first, nil })
	time.Sleep(5 * time.Millisecond)

	got, err := c.get(context.Background(), SortRecent, func(context.Context) (*presetEntry, error) {
		return nil, errors.New("postgres caído")
	})
	if err != nil || got != first {
		t.Errorf("con la base caída debería seguir sirviendo la anterior, got=%v err=%v", got, err)
	}

	// Sin versión anterior, el error sí se propaga.
	c2 := newPresetCache(time.Hour)
	if _, err := c2.get(context.Background(), SortPopular, func(context.Context) (*presetEntry, error) {
		return nil, errors.New("postgres caído")
	}); err == nil {
		t.Error("sin versión anterior debería devolver el error")
	}
}

func TestPresetCacheIgnoresUnsupportedSort(t *testing.T) {
	c := newPresetCache(time.Hour)
	e, err := c.get(context.Background(), SortAgeAsc, func(context.Context) (*presetEntry, error) {
		t.Error("no debería construir nada para un orden no cacheable")
		return nil, nil
	})
	if e != nil || err != nil {
		t.Errorf("esperaba nil,nil; got %v, %v", e, err)
	}
}
