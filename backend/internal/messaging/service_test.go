package messaging

import (
	"strings"
	"testing"
)

func TestValidateBody(t *testing.T) {
	got, err := validateBody("  Hola, ¿qué tal?  ")
	if err != nil {
		t.Fatalf("mensaje válido rechazado: %v", err)
	}
	if got != "Hola, ¿qué tal?" {
		t.Errorf("validateBody no recortó los espacios: %q", got)
	}
}

func TestValidateBodyRejectsEmpty(t *testing.T) {
	cases := []string{"", "   ", "\n\t"}
	for _, c := range cases {
		if _, err := validateBody(c); err == nil {
			t.Errorf("validateBody(%q) debería rechazar un cuerpo vacío", c)
		}
	}
}

func TestValidateBodyRejectsTooLong(t *testing.T) {
	tooLong := strings.Repeat("a", MaxBodyLength+1)
	if _, err := validateBody(tooLong); err == nil {
		t.Error("un mensaje de más de 2000 caracteres debería rechazarse")
	}
}

func TestClampPagingDefaults(t *testing.T) {
	page, pageSize := clampPaging(0, 0)
	if page != 1 {
		t.Errorf("page = %d, se esperaba 1 por defecto", page)
	}
	if pageSize != DefaultPageSize {
		t.Errorf("pageSize = %d, se esperaba %d por defecto", pageSize, DefaultPageSize)
	}
}

func TestClampPagingMaxPageSize(t *testing.T) {
	_, pageSize := clampPaging(1, 9999)
	if pageSize != MaxPageSize {
		t.Errorf("pageSize = %d, se esperaba que se recortara a %d", pageSize, MaxPageSize)
	}
}

func TestClampPagingNegativePage(t *testing.T) {
	page, _ := clampPaging(-5, 10)
	if page != 1 {
		t.Errorf("page negativa debería normalizarse a 1, se obtuvo %d", page)
	}
}
