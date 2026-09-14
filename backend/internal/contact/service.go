// Package contact implementa el formulario de contacto público (Fase
// 13): no requiere sesión, cualquiera (incluso sin cuenta) debe poder
// escribir. Reenvía el mensaje por email a la bandeja de soporte
// configurada (CONTACT_INBOX_EMAIL) usando la misma abstracción
// email.Sender que el resto de la app.
package contact

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"dating-platform/backend/internal/email"
)

const (
	MaxNameLength    = 100
	MaxMessageLength = 3000
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type Service struct {
	sender       email.Sender
	contactInbox string
}

func NewService(sender email.Sender, contactInbox string) *Service {
	return &Service{sender: sender, contactInbox: contactInbox}
}

// Send valida el mensaje y lo reenvía a la bandeja de soporte. El
// remitente (name/fromEmail) no es de fiar (nadie verifica que
// controle esa dirección): se incluye tal cual en el cuerpo para que
// soporte pueda responder, pero nunca se usa como "From" real del
// envío ni se trata como una cuenta autenticada.
func (s *Service) Send(ctx context.Context, name, fromEmail, message string) error {
	name = strings.TrimSpace(name)
	fromEmail = strings.TrimSpace(fromEmail)
	message = strings.TrimSpace(message)

	if name == "" {
		return invalidField("name", "es obligatorio")
	}
	if len([]rune(name)) > MaxNameLength {
		return invalidField("name", fmt.Sprintf("no puede superar %d caracteres", MaxNameLength))
	}
	if !emailPattern.MatchString(strings.ToLower(fromEmail)) {
		return invalidField("email", "debe ser un email válido")
	}
	if message == "" {
		return invalidField("message", "es obligatorio")
	}
	if len([]rune(message)) > MaxMessageLength {
		return invalidField("message", fmt.Sprintf("no puede superar %d caracteres", MaxMessageLength))
	}

	return s.sender.Send(ctx, email.Message{
		To:      s.contactInbox,
		Subject: "Nuevo mensaje de contacto",
		Body: fmt.Sprintf(
			"De: %s <%s>\n\n%s",
			name, fromEmail, message,
		),
	})
}
