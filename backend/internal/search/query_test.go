package search

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

var testNow = time.Date(2026, 10, 3, 15, 30, 0, 0, time.UTC)

func baseParams() Params {
	return Params{ExcludeUserID: uuid.New(), Sort: SortRecent, Page: 1, PageSize: 20}
}

// sampleFor devuelve un valor NO cero del tipo dado (para activar un filtro).
func sampleFor(t reflect.Type) reflect.Value {
	v := reflect.New(t).Elem()
	switch t.Kind() {
	case reflect.Ptr:
		v.Set(reflect.New(t.Elem()))
		v.Elem().Set(sampleFor(t.Elem()))
	case reflect.Slice:
		s := reflect.MakeSlice(t, 1, 1)
		s.Index(0).Set(sampleFor(t.Elem()))
		v.Set(s)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			v.Field(i).Set(sampleFor(t.Field(i).Type))
		}
	case reflect.String:
		v.SetString("x")
	case reflect.Int:
		v.SetInt(30)
	case reflect.Float64:
		v.SetFloat(3.5)
	case reflect.Bool:
		v.SetBool(true)
	}
	return v
}

var placeholderRe = regexp.MustCompile(`\$(\d+)`)

// placeholderSet devuelve los números de placeholder que usa un SQL.
func placeholderSet(sql string) map[int]bool {
	set := map[int]bool{}
	for _, m := range placeholderRe.FindAllStringSubmatch(sql, -1) {
		n, _ := strconv.Atoi(m[1])
		set[n] = true
	}
	return set
}

// assertExactlyUses falla si el SQL no usa EXACTAMENTE los placeholders 1..n:
// Postgres rechaza una consulta con parámetros sin usar o inexistentes.
func assertExactlyUses(t *testing.T, label, sql string, n int) {
	t.Helper()
	got := placeholderSet(sql)
	for i := 1; i <= n; i++ {
		if !got[i] {
			t.Errorf("%s: el argumento $%d no se usa en el SQL", label, i)
		}
	}
	for i := range got {
		if i > n {
			t.Errorf("%s: el SQL usa $%d pero solo hay %d argumentos", label, i, n)
		}
	}
}

func kitchenSink(sort Sort) Params {
	p := baseParams()
	p.Sort = sort
	p.Page, p.PageSize = 3, 25
	reflect.ValueOf(&p.Filters).Elem().Set(sampleFor(reflect.TypeOf(Filters{})))
	return p
}

func TestBuildSearchQuery_PlaceholdersMatchArguments(t *testing.T) {
	cases := map[string]Params{
		"sin filtros":              baseParams(),
		"todos los filtros":        kitchenSink(SortRecent),
		"todos + popular":          kitchenSink(SortPopular),
		"todos + edad ascendente":  kitchenSink(SortAgeAsc),
		"todos + edad descendente": kitchenSink(SortAgeDesc),
	}
	for name, p := range cases {
		t.Run(name, func(t *testing.T) {
			q := buildSearchQuery(p, testNow)

			assertExactlyUses(t, "consulta", q.SQL, len(q.Args))
			assertExactlyUses(t, "recuento", q.CountSQL, len(q.CountArgs))

			n := len(q.Args)
			if !strings.Contains(q.SQL, fmt.Sprintf("LIMIT $%d OFFSET $%d", n-1, n)) {
				t.Errorf("LIMIT/OFFSET deben ser los dos últimos argumentos ($%d, $%d)", n-1, n)
			}
			if q.Args[n-2] != p.PageSize || q.Args[n-1] != (p.Page-1)*p.PageSize {
				t.Errorf("paginación = %v, %v; se esperaba %d, %d", q.Args[n-2], q.Args[n-1], p.PageSize, (p.Page-1)*p.PageSize)
			}
			if len(q.CountArgs) != n-2 || strings.Contains(q.CountSQL, "LIMIT") {
				t.Errorf("el recuento no debe llevar paginación: %d argumentos (de %d), SQL:\n%s", len(q.CountArgs), n, q.CountSQL)
			}
		})
	}
}

// Cada campo de Filters tiene que alterar la consulta: si alguien añade un filtro
// y olvida cablearlo en buildSearchQuery, la búsqueda lo ignoraría en silencio.
func TestEveryFilterAffectsTheQuery(t *testing.T) {
	base := baseParams()
	baseline := buildSearchQuery(base, testNow)

	typ := reflect.TypeOf(Filters{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		t.Run(field.Name, func(t *testing.T) {
			p := base
			reflect.ValueOf(&p.Filters).Elem().Field(i).Set(sampleFor(field.Type))

			got := buildSearchQuery(p, testNow)
			if got.SQL == baseline.SQL && reflect.DeepEqual(got.Args, baseline.Args) {
				t.Errorf("el filtro %s no cambia la consulta: ¿falta cablearlo en buildSearchQuery?", field.Name)
			}
		})
	}
}

func TestBuildSearchQuery_AlwaysAppliesVisibilityAndExcludesSelf(t *testing.T) {
	visible := profiles.VisibleSQL("$1")
	for name, p := range map[string]Params{
		"sin filtros": baseParams(), "todos": kitchenSink(SortPopular),
	} {
		q := buildSearchQuery(p, testNow)
		for label, sql := range map[string]string{"consulta": q.SQL, "recuento": q.CountSQL} {
			if !strings.Contains(sql, visible) {
				t.Errorf("%s/%s: falta la regla de visibilidad compartida:\n%s", name, label, sql)
			}
			if !strings.Contains(sql, "p.user_id <> $1") {
				t.Errorf("%s/%s: falta excluir al propio usuario", name, label)
			}
		}
		if q.Args[0] != p.ExcludeUserID {
			t.Errorf("%s: $1 debe ser quien busca", name)
		}
	}
}

func TestBuildSearchQuery_Ordering(t *testing.T) {
	const popularJoin = "LEFT JOIN profile_popularity pop"
	cases := []struct {
		name       string
		sort       Sort
		online     bool
		wantOrder  string
		wantJoinOn bool
	}{
		{"recientes", SortRecent, false, "ORDER BY p.created_at DESC, p.id ASC", false},
		{"edad ascendente", SortAgeAsc, false, "ORDER BY p.birth_date DESC, p.id ASC", false},
		{"edad descendente", SortAgeDesc, false, "ORDER BY p.birth_date ASC, p.id ASC", false},
		{"populares", SortPopular, false, "ORDER BY COALESCE(pop.score, 0) DESC, p.created_at DESC, p.id ASC", true},
		{"orden desconocido cae en recientes", Sort("otro"), false, "ORDER BY p.created_at DESC, p.id ASC", false},
		{"online-now manda sobre el orden pedido", SortPopular, true, "ORDER BY u.last_active_at DESC, p.id ASC", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := baseParams()
			p.Sort = tc.sort
			p.Filters.OnlineNow = tc.online
			q := buildSearchQuery(p, testNow)

			if !strings.Contains(q.SQL, tc.wantOrder) {
				t.Errorf("falta %q en:\n%s", tc.wantOrder, q.SQL)
			}
			if got := strings.Contains(q.SQL, popularJoin); got != tc.wantJoinOn {
				t.Errorf("unión con profile_popularity = %v, se esperaba %v", got, tc.wantJoinOn)
			}
			// Todos los órdenes desempatan por id: sin eso la paginación puede repetir o perder filas.
			if !strings.Contains(q.SQL, "p.id ASC") {
				t.Error("el orden debe terminar desempatando por p.id")
			}
		})
	}
}

func TestBuildSearchQuery_SpecificClauses(t *testing.T) {
	i := func(n int) *int { return &n }
	f := func(n float64) *float64 { return &n }
	s := func(v string) *string { return &v }

	cases := []struct {
		name     string
		filters  Filters
		wantSQL  []string
		wantArgs []any // a partir de $2
	}{
		{"alturas", Filters{MinHeight: i(150), MaxHeight: i(190)},
			[]string{"p.height >= $2", "p.height <= $3"}, []any{150, 190}},
		{"texto exacto", Filters{BodyType: s("athletic"), StarSign: s("leo")},
			[]string{"p.body_type = $2", "p.star_sign = $3"}, []any{"athletic", "leo"}},
		{"listas que se solapan", Filters{Sports: []string{"yoga", "tenis"}},
			[]string{"p.sports && $2"}, []any{[]string{"yoga", "tenis"}}},
		{"hijos como máximo", Filters{MaxChildren: i(2)},
			[]string{"p.children_count <= $2"}, []any{2}},
		{"idiomas", Filters{Languages: []string{"es", "en"}},
			[]string{"pl.language_code = ANY($2)"}, []any{[]string{"es", "en"}}},
		{"género", Filters{Genders: []profiles.Gender{"female", "other"}},
			[]string{"p.gender = ANY($2)"}, []any{[]string{"female", "other"}}},
		{"interés con rango", Filters{Interests: []InterestFilter{{Key: "travel", Min: i(4), Max: i(5)}}},
			[]string{"pi.interest_key = $2 AND pi.level >= $3 AND pi.level <= $4"}, []any{"travel", 4, 5}},
		{"interés solo pertenencia", Filters{Interests: []InterestFilter{{Key: "travel"}}},
			[]string{"pi.interest_key = $2"}, []any{"travel"}},
		{"personalidad", Filters{PersonalityTraits: []PersonalityFilter{{TraitKey: "openness", Min: f(3.5)}}},
			[]string{"pts.trait_key = $2 AND pts.avg_score >= $3"}, []any{"openness", 3.5}},
		{"en línea", Filters{OnlineNow: true},
			[]string{"u.last_active_at >= now() - make_interval(secs => $2::double precision)"}, []any{float64(onlineNowWindowSeconds)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := baseParams()
			p.Filters = tc.filters
			q := buildSearchQuery(p, testNow)

			for _, want := range tc.wantSQL {
				if !strings.Contains(q.SQL, want) {
					t.Errorf("falta %q en:\n%s", want, q.SQL)
				}
			}
			if got := q.Args[1 : 1+len(tc.wantArgs)]; !reflect.DeepEqual(got, tc.wantArgs) {
				t.Errorf("argumentos = %#v, se esperaba %#v", got, tc.wantArgs)
			}
		})
	}
}

// --- Edad ---------------------------------------------------------------------------

// ageBounds debe clasificar a cada persona igual que profiles.AgeAt, que es la
// edad que ve el usuario en las tarjetas. Se comprueba sobre cada día de nacimiento
// de ~100 años y varias fechas de "hoy" conflictivas (29 de febrero incluido).
func TestAgeBounds_AgreeWithAgeAtEveryDay(t *testing.T) {
	nows := []time.Time{
		time.Date(2026, 10, 3, 15, 30, 0, 0, time.UTC),
		time.Date(2026, 2, 28, 23, 59, 0, 0, time.UTC),
		time.Date(2026, 3, 1, 0, 1, 0, 0, time.UTC),
		time.Date(2028, 2, 28, 12, 0, 0, 0, time.UTC),
		time.Date(2028, 2, 29, 12, 0, 0, 0, time.UTC), // día bisiesto
		time.Date(2028, 3, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 29, 8, 0, 0, 0, time.UTC),
		time.Date(2025, 12, 31, 23, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2027, 2, 28, 12, 0, 0, 0, time.UTC),
	}
	ages := []int{18, 19, 30, 65, 120}

	for _, now := range nows {
		first := time.Date(now.Year()-125, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		last := time.Date(now.Year()-15, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		for _, age := range ages {
			age := age
			minBirth, _ := ageBounds(now, &age, nil)
			_, maxBirth := ageBounds(now, nil, &age)

			bad := 0
			for birth := first; !birth.After(last); birth = birth.AddDate(0, 0, 1) {
				real := profiles.AgeAt(birth, now)

				if gotMin, wantMin := !birth.After(*minBirth), real >= age; gotMin != wantMin {
					bad++
					t.Errorf("hoy=%s min_age=%d: nacido %s (edad real %d): ¿incluido? %v, debería ser %v",
						now.Format("2006-01-02"), age, birth.Format("2006-01-02"), real, gotMin, wantMin)
				}
				if gotMax, wantMax := !birth.Before(*maxBirth), real <= age; gotMax != wantMax {
					bad++
					t.Errorf("hoy=%s max_age=%d: nacido %s (edad real %d): ¿incluido? %v, debería ser %v",
						now.Format("2006-01-02"), age, birth.Format("2006-01-02"), real, gotMax, wantMax)
				}
				if bad > 5 {
					t.FailNow()
				}
			}
		}
	}
}

func TestAgeBounds_NilWhenNotRequested(t *testing.T) {
	if min, max := ageBounds(testNow, nil, nil); min != nil || max != nil {
		t.Errorf("sin filtro de edad no debe haber límites: %v %v", min, max)
	}
}

func TestYearsAgo(t *testing.T) {
	d := func(y int, m time.Month, day int) time.Time { return time.Date(y, m, day, 0, 0, 0, 0, time.UTC) }
	cases := []struct {
		y    int
		m    time.Month
		day  int
		n    int
		want time.Time
	}{
		{2026, time.October, 3, 18, d(2008, time.October, 3)},
		{2028, time.February, 29, 4, d(2024, time.February, 29)},  // existe
		{2028, time.February, 29, 18, d(2010, time.February, 28)}, // no existe: último día de febrero
		{2028, time.February, 29, 1, d(2027, time.February, 28)},
		{2026, time.December, 31, 30, d(1996, time.December, 31)},
	}
	for _, tc := range cases {
		if got := yearsAgo(tc.y, tc.m, tc.day, tc.n); !got.Equal(tc.want) {
			t.Errorf("yearsAgo(%d-%02d-%02d, %d) = %s, se esperaba %s", tc.y, tc.m, tc.day, tc.n,
				got.Format("2006-01-02"), tc.want.Format("2006-01-02"))
		}
	}
}
