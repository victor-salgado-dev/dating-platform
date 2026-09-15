package auth

import "errors"

var (
	// ErrInvalidCredentials cubre tanto email inexistente como contraseña
	// incorrecta: nunca se distingue de cara al cliente, para no filtrar
	// qué emails están registrados.
	ErrInvalidCredentials = errors.New("auth: credenciales inválidas")

	// ErrAccountSuspended indica que la cuenta existe pero está suspendida
	// (moderación) y no puede iniciar sesión.
	ErrAccountSuspended = errors.New("auth: cuenta suspendida")

	// ErrTokenInvalid cubre tokens (sesión, verificación, reset) que no
	// existen, ya se usaron, o no corresponden a los datos enviados.
	ErrTokenInvalid = errors.New("auth: token inválido o caducado")

	// ErrWeakPassword se devuelve cuando una contraseña no cumple los
	// requisitos mínimos.
	ErrWeakPassword = errors.New("auth: la contraseña no cumple los requisitos mínimos")

	// ErrEmailAlreadyVerified evita reenviar verificación innecesariamente.
	ErrEmailAlreadyVerified = errors.New("auth: el email ya está verificado")

	// ErrInvalidEmail se devuelve cuando el email no tiene un formato válido.
	ErrInvalidEmail = errors.New("auth: formato de email inválido")

	// ErrTermsNotAccepted indica que el registro requiere aceptar explícitamente
	// los Términos y la Política de Privacidad.
	ErrTermsNotAccepted = errors.New("auth: debes aceptar los Términos y la Política de Privacidad")
)
