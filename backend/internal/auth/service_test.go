package auth

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got, err := normalizeEmail("  Ada@Example.COM  ")
	if err != nil {
		t.Fatalf("normalizeEmail devolvió error para un email válido: %v", err)
	}
	if got != "ada@example.com" {
		t.Errorf("normalizeEmail = %q, se esperaba %q (minúsculas, sin espacios)", got, "ada@example.com")
	}
}

func TestNormalizeEmailRejectsInvalidFormats(t *testing.T) {
	invalid := []string{"", "no-es-un-email", "falta-arroba.com", "@sin-usuario.com", "usuario@", "con espacio@example.com"}

	for _, v := range invalid {
		if _, err := normalizeEmail(v); err == nil {
			t.Errorf("normalizeEmail(%q) debería devolver error", v)
		}
	}
}

func TestMinPasswordLengthIsEnforcedByCaller(t *testing.T) {
	// MinPasswordLength es la constante que Service.Register/ResetPassword
	// usan para rechazar contraseñas débiles. Este test fija su valor
	// esperado para que un cambio accidental no pase desapercibido.
	if MinPasswordLength != 8 {
		t.Errorf("MinPasswordLength = %d, se esperaba 8", MinPasswordLength)
	}
}
