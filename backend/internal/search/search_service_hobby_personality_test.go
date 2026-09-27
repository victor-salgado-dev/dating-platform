package search

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
)

// =====================================================================
// buildParams: nivel de intereses (has_level=true)
// =====================================================================

func TestBuildParamsInterestBareKeyMeansMembershipAnyLevel(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{Interests: []string{"travel", "gardening"}})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(params.Filters.Interests) != 2 {
		t.Fatalf("se esperaban 2 filtros de interés, hay %d", len(params.Filters.Interests))
	}
	for _, itf := range params.Filters.Interests {
		if itf.Min != nil || itf.Max != nil {
			t.Errorf("interés %q sin bounds debería tener Min/Max nil, tiene %+v", itf.Key, itf)
		}
	}
}

func TestBuildParamsInterestBoundsCreateFilter(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{
		InterestBounds: map[string]RawBounds{"travel": {Min: "4"}},
	})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(params.Filters.Interests) != 1 {
		t.Fatalf("se esperaba 1 filtro de interés, hay %d", len(params.Filters.Interests))
	}
	itf := params.Filters.Interests[0]
	if itf.Key != "travel" || itf.Min == nil || *itf.Min != 4 || itf.Max != nil {
		t.Errorf("filtro construido incorrectamente: %+v", itf)
	}
}

func TestBuildParamsInterestMergesBareAndBounds(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{
		Interests:      []string{"travel"},
		InterestBounds: map[string]RawBounds{"travel": {Min: "3"}},
	})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(params.Filters.Interests) != 1 {
		t.Fatalf("la misma clave en ambas fuentes no debería duplicar el filtro, hay %d", len(params.Filters.Interests))
	}
	if params.Filters.Interests[0].Min == nil || *params.Filters.Interests[0].Min != 3 {
		t.Errorf("el límite de InterestBounds debería haberse aplicado: %+v", params.Filters.Interests[0])
	}
}

func TestBuildParamsMultipleInterestFiltersAreOrderedDeterministically(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{
		InterestBounds: map[string]RawBounds{
			"travel":    {Min: "3"},
			"gardening": {Max: "3"},
		},
	})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(params.Filters.Interests) != 2 {
		t.Fatalf("se esperaban 2 filtros, hay %d", len(params.Filters.Interests))
	}
	if params.Filters.Interests[0].Key != "gardening" || params.Filters.Interests[1].Key != "travel" {
		t.Errorf("se esperaba orden alfabético (gardening, travel), se obtuvo: %+v", params.Filters.Interests)
	}
}

func TestBuildParamsRejectsInterestLevelOutOfRange(t *testing.T) {
	for _, bad := range []string{"0", "6", "-1"} {
		_, err := buildParams(uuid.New(), RawQuery{
			InterestBounds: map[string]RawBounds{"travel": {Min: bad}},
		})
		if err == nil {
			t.Errorf("nivel %q fuera de 1-5 debería rechazarse", bad)
		}
	}
}

func TestBuildParamsRejectsInterestMinGreaterThanMax(t *testing.T) {
	_, err := buildParams(uuid.New(), RawQuery{
		InterestBounds: map[string]RawBounds{"travel": {Min: "4", Max: "2"}},
	})
	if err == nil {
		t.Error("interest_travel_min > interest_travel_max debería rechazarse")
	}
}

func TestBuildParamsRejectsNonNumericInterestBound(t *testing.T) {
	_, err := buildParams(uuid.New(), RawQuery{
		InterestBounds: map[string]RawBounds{"travel": {Min: "mucho"}},
	})
	if err == nil {
		t.Error("un límite de interés no numérico debería rechazarse")
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
// parseKeyedBounds (usada por el Handler para leer ?trait_x_min= /
// ?trait_x_max= de la query string). Es un parser genérico por prefijo,
// así que se prueba aquí con un prefijo neutro además de "trait_" para
// dejar claro que no depende de ningún dominio concreto (ya no existen
// los hobbies, que antes usaban este mismo parser con "hobby_").
// =====================================================================

func TestParseKeyedBounds(t *testing.T) {
	q := url.Values{
		"opt_meeting_new_people_max": {"2"}, // clave con guiones bajos: no debe romper el parseo
		"opt_min_length_min":         {"4"},
		"trait_openness_min":         {"3"},
		"page":                       {"2"}, // no debe colarse
		"opt_empty_min":              {""},  // vacío: se ignora
	}

	optBounds := parseKeyedBounds(q, "opt_")
	if len(optBounds) != 2 {
		t.Fatalf("se esperaban 2 claves con prefijo 'opt_', hubo %d: %+v", len(optBounds), optBounds)
	}
	if optBounds["min_length"].Min != "4" {
		t.Errorf("opt_min_length_min no se leyó bien: %+v", optBounds["min_length"])
	}
	if optBounds["meeting_new_people"].Max != "2" {
		t.Errorf("una clave con guiones bajos no se parseó bien: %+v", optBounds["meeting_new_people"])
	}

	traitBounds := parseKeyedBounds(q, "trait_")
	if len(traitBounds) != 1 || traitBounds["openness"].Min != "3" {
		t.Errorf("trait_openness_min no se leyó bien: %+v", traitBounds)
	}
}

func TestParseKeyedBoundsMergesMinAndMaxForSameKey(t *testing.T) {
	q := url.Values{
		"trait_openness_min": {"3"},
		"trait_openness_max": {"5"},
	}
	bounds := parseKeyedBounds(q, "trait_")
	if len(bounds) != 1 {
		t.Fatalf("min y max de la misma clave deberían fundirse en una entrada, hay %d", len(bounds))
	}
	b := bounds["openness"]
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
	if bounds := parseKeyedBounds(q, "opt_"); len(bounds) != 0 {
		t.Errorf("no debería haber coincidencias de 'opt_', hubo: %+v", bounds)
	}
	if bounds := parseKeyedBounds(q, "trait_"); len(bounds) != 0 {
		t.Errorf("'trait_foo' sin sufijo _min/_max no debería generar ninguna entrada, hubo: %+v", bounds)
	}
}
