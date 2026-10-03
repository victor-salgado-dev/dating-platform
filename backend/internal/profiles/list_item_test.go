package profiles

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// fakeRow rellena los destinos de Scan por reflexión y comprueba que el número
// de columnas coincide con el de destinos (como haría pgx).
type fakeRow struct {
	values []any
	err    error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.values) {
		return fmt.Errorf("%d destinos para %d columnas", len(dest), len(r.values))
	}
	for i, d := range dest {
		reflect.ValueOf(d).Elem().Set(reflect.ValueOf(r.values[i]))
	}
	return nil
}

func baseColumns(id uuid.UUID, birth time.Time) []any {
	region := "Madrid"
	return []any{id, "Ana", birth, "female", "ES", &region, true}
}

func TestScanBaseListItem(t *testing.T) {
	id := uuid.New()
	birth := time.Now().AddDate(-30, 0, -10) // 30 años cumplidos
	likedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	var extra time.Time
	item, err := ScanBaseListItem(fakeRow{values: append(baseColumns(id, birth), likedAt)}, &extra)
	if err != nil {
		t.Fatalf("ScanBaseListItem: %v", err)
	}

	if item.ProfileID != id || item.DisplayName != "Ana" || item.CountryCode != "ES" || !item.HasPhoto {
		t.Errorf("columnas base mal escaneadas: %+v", item)
	}
	if item.Region == nil || *item.Region != "Madrid" {
		t.Errorf("region = %v", item.Region)
	}
	if item.Gender != GenderFemale || item.Age != 30 {
		t.Errorf("gender=%q age=%d, se esperaba female/30", item.Gender, item.Age)
	}
	if !extra.Equal(likedAt) {
		t.Errorf("el extra se escanea tras las 7 columnas base: %v", extra)
	}
	if item.Liked || item.Favorited || item.ReceivedLike || item.ReceivedFavorite {
		t.Error("ScanBaseListItem no rellena los flags de interacción")
	}
}

func TestScanListItem_FlagsComeAfterExtras(t *testing.T) {
	id := uuid.New()
	var extra time.Time
	var total int
	at := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	// 7 base + 2 extras + 4 flags (liked, favorited, received_like, received_favorite)
	cols := append(baseColumns(id, time.Now().AddDate(-25, -1, 0)), at, 42, true, false, true, false)
	item, err := ScanListItem(fakeRow{values: cols}, &extra, &total)
	if err != nil {
		t.Fatalf("ScanListItem: %v", err)
	}

	if !extra.Equal(at) || total != 42 {
		t.Errorf("extras = %v / %d", extra, total)
	}
	if !item.Liked || item.Favorited || !item.ReceivedLike || item.ReceivedFavorite {
		t.Errorf("flags mal asignados: liked=%v favorited=%v received_like=%v received_favorite=%v",
			item.Liked, item.Favorited, item.ReceivedLike, item.ReceivedFavorite)
	}
}

func TestScanListItem_PropagatesRowError(t *testing.T) {
	boom := errors.New("fallo de red")
	for name, scan := range map[string]func() (BaseListItem, error){
		"base":  func() (BaseListItem, error) { return ScanBaseListItem(fakeRow{err: boom}) },
		"flags": func() (BaseListItem, error) { return ScanListItem(fakeRow{err: boom}) },
	} {
		item, err := scan()
		if !errors.Is(err, boom) {
			t.Errorf("%s: err = %v", name, err)
		}
		if item.ProfileID != uuid.Nil || item.DisplayName != "" {
			t.Errorf("%s: con error debe devolver el valor cero, no %+v", name, item)
		}
	}
}

func TestViewerFlagsSQL(t *testing.T) {
	sql := ViewerFlagsSQL("me", "p")
	order := []string{"AS liked", "AS favorited", "AS received_like", "AS received_favorite"}
	pos := -1
	for _, col := range order {
		i := strings.Index(sql, col)
		if i <= pos {
			t.Fatalf("columna %q ausente o fuera de orden en:\n%s", col, sql)
		}
		pos = i
	}
	for _, want := range []string{"fl.from_profile_id = me.id AND fl.to_profile_id = p.id", "rf.user_id = p.user_id AND rf.favorite_profile_id = me.id"} {
		if !strings.Contains(sql, want) {
			t.Errorf("falta %q en el SQL", want)
		}
	}
}

func TestViewerFlagsSQL_RejectsUnsafeOrReservedAliases(t *testing.T) {
	for _, alias := range []string{"", "1abc", "me; DROP TABLE users", "a b", `x"`, "a.b", "fl", "ff", "rl", "rf"} {
		for _, args := range [][2]string{{alias, "p"}, {"me", alias}} {
			func() {
				defer func() {
					if recover() == nil {
						t.Errorf("ViewerFlagsSQL(%q, %q) debería entrar en pánico", args[0], args[1])
					}
				}()
				ViewerFlagsSQL(args[0], args[1])
			}()
		}
	}
}
