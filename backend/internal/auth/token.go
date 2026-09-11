package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// generateToken crea un token opaco criptográficamente seguro, apto para
// incluir en una URL (verificación de email, reset de contraseña).
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: no se pudo generar token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashToken calcula el hash del token para almacenarlo en Redis.
// Nunca guardamos el token en claro: si Redis quedara expuesto, un
// atacante no podría reutilizar los tokens ya emitidos.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
