package search

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
)

// =====================================================================
// buildParams: hobbies
// =====================================================================

func TestBuildParamsHobbyBareKeyMeansLikedAnyIntensity(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{Hobbies: []string{"traveling", "gardening"}})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(params.Filters.Hobbies) != 2 {
		t.Fatalf("se esperaban 2 filtros de hobby, hay %d", len(params.Filters.Hobbies))
	}
	for _, hf := range params.Filters.Hobbies {
		if hf.Min != nil || hf.Max != nil {
			t.Errorf("hobby %q sin bounds debería tener Min/Max nil, tiene %+v", hf.Key, hf)
		}
	}
}

func TestBuildParamsHobbyBoundsCreateFilter(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{
		HobbyBounds: map[string]RawBounds{"traveling": {Min: "4"}},
	})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(params.Filters.Hobbies) != 1 {
		t.Fatalf("se esperaba 1 filtro de hobby, hay %d", len(params.Filters.Hobbies))
	}
	hf := params.Filters.Hobbies[0]
	if hf.Key != "traveling" || hf.Min == nil || *hf.Min != 4 || hf.Max != nil {
		t.Errorf("filtro construido incorrectamente: %+v", hf)
	}
}

func TestBuildParamsHobbyMergesBareAndBounds(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{
		Hobbies:     []string{"traveling"},
		HobbyBounds: map[string]RawBounds{"traveling": {Min: "3"}},
	})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(params.Filters.Hobbies) != 1 {
		t.Fatalf("la misma clave en ambas fuentes no debería duplicar el filtro, hay %d", len(params.Filters.Hobbies))
	}
	if params.Filters.Hobbies[0].Min == nil || *params.Filters.Hobbies[0].Min != 3 {
		t.Errorf("el límite de HobbyBounds debería haberse aplicado: %+v", params.Filters.Hobbies[0])
	}
}

func TestBuildParamsMultipleHobbyFiltersAreOrderedDeterministically(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{
		HobbyBounds: map[string]RawBounds{
			"traveling": {Min: "3"},
			"gardening": {Max: "3"},
		},
	})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(params.Filters.Hobbies) != 2 {
		t.Fatalf("se esperaban 2 filtros, hay %d", len(params.Filters.Hobbies))
	}
	if params.Filters.Hobbies[0].Key != "gardening" || params.Filters.Hobbies[1].Key != "traveling" {
		t.Errorf("se esperaba orden alfabético (gardening, traveling), se obtuvo: %+v", params.Filters.Hobbies)
	}
}

func TestBuildParamsRejectsHobbyIntensityOutOfRange(t *testing.T) {
	for _, bad := range []string{"0", "6", "-1"} {
		_, err := buildParams(uuid.New(), RawQuery{
			HobbyBounds: map[string]RawBounds{"traveling": {Min: bad}},
		})
		if err == nil {
			t.Errorf("intensidad %q fuera de 1-5 debería rechazarse", bad)
		}
	}
}

func TestBuildParamsRejectsHobbyMinGreaterThanMax(t *testing.T) {
	_, err := buildParams(uuid.New(), RawQuery{
		HobbyBounds: map[string]RawBounds{"traveling": {Min: "4", Max: "2"}},
	})
	if err == nil {
		t.Error("hobby_traveling_min > hobby_traveling_max debería rechazarse")
	}
}

func TestBuildParamsRejectsNonNumericHobbyBound(t *testing.T) {
	_, err := buildParams(uuid.New(), RawQuery{
		HobbyBounds: map[string]RawBounds{"traveling": {Min: "mucho"}},
	})
	if err == nil {
		t.Error("un límite de hobby no numérico debería rechazarse")
	}
}

// =====================================================================
// buildParams: personalidad
// =====================================================================

func TestBuildParamsPersonalityTraitValid(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{
		PersonalityTraitBounds: map[string]RawBounds{"extraversion": {Min: "3.5"}},
	})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(params.Filters.PersonalityTraits) != 1 {
		t.Fatalf("se esperaba 1 filtro de personalidad, hay %d", len(params.Filters.PersonalityTraits))
	}
	pf := params.Filters.PersonalityTraits[0]
	if pf.TraitKey != "extraversion" || pf.Min == nil || *pf.Min != 3.5 {
		t.Errorf("filtro construido incorrectamente: %+v", pf)
	}
}

func TestBuildParamsRejectsInvalidPersonalityTrait(t *testing.T) {
	_, err := buildParams(uuid.New(), RawQuery{
		PersonalityTraitBounds: map[string]RawBounds{"charisma": {Min: "3"}},
	})
	if err == nil {
		t.Error("un rasgo de personalidad que no existe debería rechazarse")
	}
}

func TestBuildParamsRejectsPersonalityScoreOutOfRange(t *testing.T) {
	_, err := buildParams(uuid.New(), RawQuery{
		PersonalityTraitBounds: map[string]RawBounds{"openness": {Max: "9"}},
	})
	if err == nil {
		t.Error("una puntuación fuera de 1-5 debería rechazarse")
	}
}

func TestBuildParamsRejectsPersonalityMinGreaterThanMax(t *testing.T) {
	_, err := buildParams(uuid.New(), RawQuery{
		PersonalityTraitBounds: map[string]RawBounds{"openness": {Min: "4", Max: "2"}},
	})
	if err == nil {
		t.Error("trait_openness_min > trait_openness_max debería rechazarse")
	}
}

func TestBuildParamsNoPersonalityBoundsProducesNoFilter(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(params.Filters.PersonalityTraits) != 0 {
		t.Errorf("sin bounds no debería generarse ningún filtro de personalidad, hay %d", len(params.Filters.PersonalityTraits))
	}
}

// =====================================================================
// parseKeyedBounds (usada por el Handler para leer ?hobby_x_min= y
// ?trait_x_max= de la query string)
// =====================================================================

func TestParseKeyedBounds(t *testing.T) {
	q := url.Values{
		"hobby_traveling_min":          {"4"},
		"hobby_meeting_new_people_max": {"2"}, // clave con guiones bajos: no debe romper el parseo
		"trait_openness_min":           {"3"},
		"page":                         {"2"}, // no debe colarse
		"hobby_empty_min":              {""},  // vacío: se ignora
	}

	hobbyBounds := parseKeyedBounds(q, "hobby_")
	if len(hobbyBounds) != 2 {
		t.Fatalf("se esperaban 2 claves de hobby, hubo %d: %+v", len(hobbyBounds), hobbyBounds)
	}
	if hobbyBounds["traveling"].Min != "4" {
		t.Errorf("hobby_traveling_min no se leyó bien: %+v", hobbyBounds["traveling"])
	}
	if hobbyBounds["meeting_new_people"].Max != "2" {
		t.Errorf("una clave con guiones bajos no se parseó bien: %+v", hobbyBounds["meeting_new_people"])
	}

	traitBounds := parseKeyedBounds(q, "trait_")
	if len(traitBounds) != 1 || traitBounds["openness"].Min != "3" {
		t.Errorf("trait_openness_min no se leyó bien: %+v", traitBounds)
	}
}

func TestParseKeyedBoundsMergesMinAndMaxForSameKey(t *testing.T) {
	q := url.Values{
		"hobby_traveling_min": {"3"},
		"hobby_traveling_max": {"5"},
	}
	bounds := parseKeyedBounds(q, "hobby_")
	if len(bounds) != 1 {
		t.Fatalf("min y max de la misma clave deberían fundirse en una entrada, hay %d", len(bounds))
	}
	b := bounds["traveling"]
	if b.Min != "3" || b.Max != "5" {
		t.Errorf("min/max no se fundieron bien: %+v", b)
	}
}

func TestParseKeyedBoundsIgnoresUnrelatedParams(t *testing.T) {
	q := url.Values{
		"gender":    {"female"},
		"min_age":   {"25"},
		"sort":      {"recent"},
		"trait_foo": {"3"}, // sin sufijo _min/_max: se ignora
	}
	if bounds := parseKeyedBounds(q, "hobby_"); len(bounds) != 0 {
		t.Errorf("no debería haber coincidencias de 'hobby_', hubo: %+v", bounds)
	}
	if bounds := parseKeyedBounds(q, "trait_"); len(bounds) != 0 {
		t.Errorf("'trait_foo' sin sufijo _min/_max no debería generar ninguna entrada, hubo: %+v", bounds)
	}
}
