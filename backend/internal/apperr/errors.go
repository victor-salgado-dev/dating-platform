// Package apperr reúne los tipos de error que comparten varios dominios.
// Antes cada paquete declaraba su propio ValidationError idéntico; ahora
// todos apuntan a este vía un alias, así que errors.As sigue funcionando
// igual pero solo hay una definición que mantener.
package apperr

import "fmt"

// ValidationError señala que un campo no cumple las reglas de negocio
// (formato, rango, valor permitido). El handler HTTP la traduce a un 400
// incluyendo el nombre del campo.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("campo %q inválido: %s", e.Field, e.Message)
}

// InvalidField construye un *ValidationError. Atajo para no repetir el
// literal en cada regla de validación.
func InvalidField(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
