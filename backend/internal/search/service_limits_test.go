package search

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestBuildParams_InterestFilterCap(t *testing.T) {
	keys := func(n int) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = fmt.Sprintf("interes_%02d", i)
		}
		return out
	}

	if _, err := buildParams(uuid.New(), RawQuery{Interests: keys(MaxInterestFilters)}); err != nil {
		t.Errorf("%d intereses (el máximo) deben aceptarse: %v", MaxInterestFilters, err)
	}
	if _, err := buildParams(uuid.New(), RawQuery{Interests: keys(MaxInterestFilters + 1)}); err == nil {
		t.Errorf("%d intereses deben rechazarse", MaxInterestFilters+1)
	}

	// Los límites de nivel cuentan igual que los intereses sueltos, y una clave en
	// ambos sitios (o repetida) es UN solo filtro.
	bounds := map[string]RawBounds{}
	for _, k := range keys(MaxInterestFilters + 1) {
		bounds[k] = RawBounds{Min: "3"}
	}
	if _, err := buildParams(uuid.New(), RawQuery{InterestBounds: bounds}); err == nil {
		t.Error("el tope también debe aplicarse a los filtros con rango de nivel")
	}
	dup := RawQuery{Interests: append(keys(MaxInterestFilters), keys(MaxInterestFilters)...), InterestBounds: map[string]RawBounds{"interes_00": {Min: "2"}}}
	if _, err := buildParams(uuid.New(), dup); err != nil {
		t.Errorf("las claves repetidas no deben contar dos veces: %v", err)
	}
}

func TestBuildParams_PageCap(t *testing.T) {
	cases := map[string]bool{
		"1": true, "2": true, fmt.Sprint(MaxPage): true,
		fmt.Sprint(MaxPage + 1): false,
		"99999999999999999999":  false, // desborda int
		"9223372036854775807":   false, // desbordaría el OFFSET
		"0":                     false, "-1": false, "x": false,
	}
	for page, ok := range cases {
		params, err := buildParams(uuid.New(), RawQuery{Page: page})
		if (err == nil) != ok {
			t.Errorf("page=%q: err=%v, aceptada esperada=%v", page, err, ok)
		}
		if ok && fmt.Sprint(params.Page) != page {
			t.Errorf("page=%q: Page=%d", page, params.Page)
		}
	}
}

// Un texto en blanco no es un filtro, como en el resto de parámetros de texto
// (antes la nacionalidad "   " filtraba por la cadena vacía y no devolvía nada).
func TestBuildParams_BlankNationalityIsNoFilter(t *testing.T) {
	for _, blank := range []string{"", " ", "   \t"} {
		params, err := buildParams(uuid.New(), RawQuery{Nationality: blank})
		if err != nil {
			t.Fatalf("nationality=%q: %v", blank, err)
		}
		if params.Filters.Nationality != nil {
			t.Errorf("nationality=%q debe quedar sin filtrar, es %q", blank, *params.Filters.Nationality)
		}
	}
	params, _ := buildParams(uuid.New(), RawQuery{Nationality: " de "})
	if params.Filters.Nationality == nil || *params.Filters.Nationality != "DE" {
		t.Errorf("nationality válida: %v", params.Filters.Nationality)
	}
}

// Antes los límites inválidos se validaban recorriendo un map, o sea en orden al
// azar: con dos intereses mal formados el error cambiaba de una petición a otra.
func TestBuildParams_InterestBoundErrorsAreDeterministic(t *testing.T) {
	raw := RawQuery{InterestBounds: map[string]RawBounds{
		"yoga":    {Min: "9"},
		"cocina":  {Min: "x"},
		"viajes":  {Min: "0"},
		"lectura": {Min: "5", Max: "1"},
	}}
	_, first := buildParams(uuid.New(), raw)
	if first == nil || !strings.Contains(first.Error(), "interest_cocina_min") {
		t.Fatalf("debe informar del primero por orden alfabético (cocina): %v", first)
	}
	for i := 0; i < 100; i++ {
		if _, err := buildParams(uuid.New(), raw); err == nil || err.Error() != first.Error() {
			t.Fatalf("el error cambió entre ejecuciones: %v / %v", first, err)
		}
	}
}

// El límite de un rasgo se busca por la clave original; antes, si traía espacios,
// se recortaba la clave y NO se encontraba su límite (el filtro quedaba sin rango).
func TestBuildParams_TraitBoundsSurviveWhitespaceInKey(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{
		PersonalityTraitBounds: map[string]RawBounds{" openness ": {Min: "4"}},
	})
	if err != nil {
		t.Fatalf("buildParams: %v", err)
	}
	if len(params.Filters.PersonalityTraits) != 1 {
		t.Fatalf("filtros de personalidad = %d", len(params.Filters.PersonalityTraits))
	}
	if pf := params.Filters.PersonalityTraits[0]; pf.TraitKey != "openness" || pf.Min == nil || *pf.Min != 4 {
		t.Errorf("el límite se perdió: %+v", pf)
	}
}
