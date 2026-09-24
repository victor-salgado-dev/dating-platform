package search

import (
	"testing"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

func TestBuildParamsDefaults(t *testing.T) {
	me := uuid.New()

	params, err := buildParams(me, RawQuery{})
	if err != nil {
		t.Fatalf("buildParams con query vacía no debería fallar: %v", err)
	}

	if params.ExcludeUserID != me {
		t.Error("ExcludeUserID debería ser quien busca, para no aparecer en sus propios resultados")
	}
	if params.Sort != SortRecent {
		t.Errorf("Sort por defecto = %q, se esperaba %q", params.Sort, SortRecent)
	}
	if params.Page != 1 {
		t.Errorf("Page por defecto = %d, se esperaba 1", params.Page)
	}
	if params.PageSize != DefaultPageSize {
		t.Errorf("PageSize por defecto = %d, se esperaba %d", params.PageSize, DefaultPageSize)
	}
}

func TestBuildParamsClampsPageSize(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{PageSize: "9999"})
	if err != nil {
		t.Fatalf("page_size grande no debería ser un error, solo recortarse: %v", err)
	}
	if params.PageSize != MaxPageSize {
		t.Errorf("PageSize = %d, se esperaba que se recortara a %d", params.PageSize, MaxPageSize)
	}
}

func TestBuildParamsRejectsInvalidGender(t *testing.T) {
	_, err := buildParams(uuid.New(), RawQuery{Genders: []string{"robot"}})
	if err == nil {
		t.Error("un género que no existe debería rechazarse")
	}
}

func TestBuildParamsRejectsMinAgeGreaterThanMaxAge(t *testing.T) {
	_, err := buildParams(uuid.New(), RawQuery{MinAge: "40", MaxAge: "30"})
	if err == nil {
		t.Error("min_age > max_age debería rechazarse")
	}
}

func TestBuildParamsRejectsAgeOutOfRange(t *testing.T) {
	_, err := buildParams(uuid.New(), RawQuery{MinAge: "10"})
	if err == nil {
		t.Error("min_age por debajo de 18 debería rechazarse (sección 1: V1 solo mayores de edad)")
	}

	_, err = buildParams(uuid.New(), RawQuery{MaxAge: "200"})
	if err == nil {
		t.Error("max_age fuera de rango debería rechazarse")
	}
}

func TestBuildParamsRejectsInvalidSort(t *testing.T) {
	_, err := buildParams(uuid.New(), RawQuery{Sort: "aleatorio"})
	if err == nil {
		t.Error("un valor de sort que no existe debería rechazarse")
	}
}

func TestBuildParamsAcceptsKnownRelationshipGoal(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{RelationshipGoals: []string{string(profiles.RelationshipLongTerm)}})
	if err != nil {
		t.Fatalf("un objetivo de relación válido no debería rechazarse: %v", err)
	}
	if len(params.Filters.RelationshipGoals) != 1 || params.Filters.RelationshipGoals[0] != profiles.RelationshipLongTerm {
		t.Error("el filtro de objetivo de relación no se aplicó correctamente")
	}
}

// TestBuildParamsLeavesUnsetFiltersNil comprueba, a nivel de construcción
// de parámetros, la mitad "de aplicación" de la regla de los datos
// faltantes: un filtro que no se envía debe quedar en nil, no con un
// valor por defecto que luego el repositorio pudiera confundir con
// "el usuario pidió explícitamente esto". La otra mitad (que un nil
// no excluya perfiles sin ese dato) se comprueba en el test de
// integración TestSearch_MissingDataRule.
func TestBuildParamsLeavesUnsetFiltersNil(t *testing.T) {
	params, err := buildParams(uuid.New(), RawQuery{})
	if err != nil {
		t.Fatalf("buildParams con query vacía no debería fallar: %v", err)
	}

	f := params.Filters
	if f.MinAge != nil || f.MaxAge != nil || f.CountryCode != nil ||
		f.HasChildren != nil || f.WantsChildren != nil {
		t.Error("los filtros no enviados deberían quedar en nil, nunca con un valor por defecto")
	}
	if len(f.Genders) != 0 || len(f.Languages) != 0 || len(f.Interests) != 0 || len(f.RelationshipGoals) != 0 {
		t.Error("las listas de filtros no enviadas deberían quedar vacías")
	}
}
