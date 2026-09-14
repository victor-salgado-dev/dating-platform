package auth

import "testing"

func TestGenerateTokenIsUniqueAndNonEmpty(t *testing.T) {
	a, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken devolvió error: %v", err)
	}
	b, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken devolvió error: %v", err)
	}

	if a == "" || b == "" {
		t.Fatal("generateToken no debería devolver una cadena vacía")
	}
	if a == b {
		t.Error("dos tokens generados por separado no deberían coincidir")
	}
}

func TestHashTokenIsDeterministicButNotReversible(t *testing.T) {
	h1 := hashToken("mi-token-secreto")
	h2 := hashToken("mi-token-secreto")
	if h1 != h2 {
		t.Error("hashToken debería ser determinista: el mismo token siempre da el mismo hash")
	}

	h3 := hashToken("otro-token-distinto")
	if h1 == h3 {
		t.Error("tokens distintos no deberían producir el mismo hash")
	}

	if h1 == "mi-token-secreto" {
		t.Error("hashToken no debería devolver el token en claro")
	}
}
