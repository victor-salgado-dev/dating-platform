package admin

import (
	"errors"
	"fmt"
)

// ErrCannotActOnSelf evita que un admin se suspenda a sí mismo por
// error (y se quede sin poder deshacerlo).
var ErrCannotActOnSelf = errors.New("admin: no puedes realizar esta acción sobre tu propia cuenta")

// ValidationError señala un parámetro de administración inválido
// (p. ej. un estado de resolución de reporte que no existe).
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("admin: campo %q inválido: %s", e.Field, e.Message)
}

func invalidField(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
