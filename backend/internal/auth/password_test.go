package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("una-contraseña-segura")
	if err != nil {
		t.Fatalf("HashPassword devolvió error: %v", err)
	}
	if hash == "" || hash == "una-contraseña-segura" {
		t.Fatal("HashPassword debería devolver un hash, no la contraseña en claro ni una cadena vacía")
	}

	if !VerifyPassword(hash, "una-contraseña-segura") {
		t.Error("VerifyPassword debería aceptar la contraseña correcta")
	}
	if VerifyPassword(hash, "otra-contraseña") {
		t.Error("VerifyPassword no debería aceptar una contraseña incorrecta")
	}
}

func TestHashPasswordIsSalted(t *testing.T) {
	h1, err := HashPassword("misma-contraseña")
	if err != nil {
		t.Fatalf("HashPassword devolvió error: %v", err)
	}
	h2, err := HashPassword("misma-contraseña")
	if err != nil {
		t.Fatalf("HashPassword devolvió error: %v", err)
	}

	if h1 == h2 {
		t.Error("dos hashes de la misma contraseña no deberían coincidir (bcrypt incluye salt aleatorio)")
	}
}
