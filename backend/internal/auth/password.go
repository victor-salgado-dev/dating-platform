package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// MinPasswordLength es la longitud mínima exigida a una contraseña nueva.
// Reglas de complejidad más elaboradas podrán añadirse más adelante sin
// romper compatibilidad con contraseñas ya almacenadas (son hashes).
const MinPasswordLength = 8

// HashPassword genera un hash bcrypt seguro de una contraseña en claro.
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("auth: no se pudo hashear la contraseña: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword comprueba si una contraseña en claro coincide con un
// hash bcrypt almacenado.
func VerifyPassword(hash, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
