package reports

import (
	"errors"
	"fmt"
)

// ErrCannotReportSelf: no tiene sentido reportarte a ti mismo.
var ErrCannotReportSelf = errors.New("reports: no puedes reportarte a ti mismo")

// ValidationError señala un campo de reporte inválido (motivo no
// permitido, descripción demasiado larga).
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("reports: campo %q inválido: %s", e.Field, e.Message)
}

func invalidField(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
