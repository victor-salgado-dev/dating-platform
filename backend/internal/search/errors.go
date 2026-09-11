package search

import "fmt"

// ValidationError señala un parámetro de búsqueda inválido. El handler
// HTTP la traduce a un 400 con el nombre del parámetro incluido.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("search: parámetro %q inválido: %s", e.Field, e.Message)
}

func invalidParam(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
