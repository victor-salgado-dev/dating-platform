package messaging

import (
	"errors"
	"fmt"
)

var (
	// ErrConversationNotFound cubre tanto "no existe" como "existe pero
	// no eres participante": ambos casos dan el mismo 404, para no
	// filtrar si una conversación existe entre otras dos personas.
	ErrConversationNotFound = errors.New("messaging: conversación no encontrada")

	// ErrCannotMessageSelf: no tiene sentido escribirte a ti mismo.
	ErrCannotMessageSelf = errors.New("messaging: no puedes enviarte un mensaje a ti mismo")
)

// ValidationError señala un mensaje inválido (vacío, demasiado largo).
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("messaging: campo %q inválido: %s", e.Field, e.Message)
}

func invalidField(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
