package contact

import "fmt"

// ValidationError señala un campo inválido del formulario de contacto.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("contact: campo %q inválido: %s", e.Field, e.Message)
}

func invalidField(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
